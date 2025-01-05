package bootstrapping

import (
	"bytes"
	"context"
	"log"
	"sabot/lib/database"
	"sabot/lib/notify"
	"sabot/lib/pir"
	"sabot/lib/util"
	pb "sabot/proto/bootstrapping"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

var (
	localTestPrefix = ""
)

type ServerInfo struct {
	Addr        []string
	GrpcClients []*pb.BootstrappingClient
	Conns       []*grpc.ClientConn
}

type Client struct {
	*Experiment
	Id        []byte // client's identifier in DB
	Pps       []*database.DBParams
	Dpfs      []*pir.DpfClient
	NumServer int // number of servers (= 2)
	DpfSeed   int64
	Contacts  *[]database.IKVElement
	*ServerInfo
}

func InitClient(config *Config, sInfo *ServerInfo) *Client {
	c := Client{}
	c.Experiment = NewExperiment(config)
	c.NumServer = len(sInfo.Addr)
	c.DpfSeed = util.DPF_SEED

	// Set up  server infos and connections
	c.ServerInfo = sInfo
	c.GrpcClients = make([]*pb.BootstrappingClient, c.NumServer)
	c.Conns = make([]*grpc.ClientConn, c.NumServer)

	// true = client running this function
	creds, err := util.LoadTLSCred(localTestPrefix+util.CERT_C_PATH_PRE, localTestPrefix+util.CERT_CA_PATH, true)
	if err != nil {
		log.Fatalln("failed loading TLS credentials:", err)
	}

	var wg sync.WaitGroup
	wg.Add(c.NumServer)

	for i := 0; i < int(c.NumServer); i++ {
		go initWorker(&c, &wg, i, &creds)
	}
	wg.Wait()

	if c.Contacts == nil || c.Pps == nil {
		log.Fatalln("client init failed")
	}

	c.Dpfs = make([]*pir.DpfClient, len(c.Pps))
	for i, pp := range c.Pps {
		c.Dpfs[i] = pir.InitPIRClient(&database.StaticDBParams{NRows: int(pp.NRows)}, pir.RandSource())
	}

	return &c
}

func initWorker(c *Client, wg *sync.WaitGroup, i int, creds *credentials.TransportCredentials) {
	defer wg.Done()
	var err error
	c.ServerInfo.Conns[i], err = grpc.Dial(c.Addr[i],
		grpc.WithTransportCredentials(*creds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(util.MAX_MSG_SIZE), grpc.MaxCallSendMsgSize(util.MAX_MSG_SIZE)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pb.NewBootstrappingClient(c.ServerInfo.Conns[i])
	c.GrpcClients[i] = &client
	clientDeadline := time.Now().Add(util.TIMEOUT)
	ctx, cancel := context.WithTimeout(context.Background(), util.TIMEOUT)
	ctx, cancel = context.WithDeadline(ctx, clientDeadline)
	defer cancel()
	conf := &pb.Config{
		ResetServer: c.Config.ResetServer,
		Dbfile:      c.Config.Dbfile,
		MultiClient: c.Config.MultiClient,
		NumThreads:  c.Config.NumThreads,
		CIdx:        c.Idx,
		NumTargets:  c.RateS,
		DbType:      util.Uint32ToByteSlice(uint32(c.Config.DBType)),
	}

	res, err := (*c.GrpcClients[i]).SetupExperiment(ctx, conf)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// we only need to do this once in our benchmark setup
	if i == 0 && len(res.Targets) == int(util.KEY_LENGTH*conf.NumTargets) {

		c.Pps = make([]*database.DBParams, c.NumServer)
		c.Id = res.CKW
		for j, pp := range res.Params {
			c.Pps[j] = &database.DBParams{
				NRows:              pp.Nrows,
				Auth:               pp.Auth,
				Seed:               pp.Seed,
				SegmentLength:      pp.SegLen,
				SegmentLengthMask:  pp.SegLenMask,
				SegmentCount:       pp.SegCount,
				SegmentCountLength: pp.SegCountLen,
				KeyLength:          pp.KeyLen,
				ValueLength:        pp.ValLen,
				ProofLen:           pp.ProofLen,
				Root:               pp.Root,
				RecordLength:       pp.RecLength,
			}
		}

		// save keywords for receivers
		receiver := make([]database.IKVElement, c.RateS)
		for j := range receiver {
			receiver[j].Key = res.Targets[j*int(util.KEY_LENGTH) : (j+1)*int(util.KEY_LENGTH)]
		}
		c.Contacts = &receiver

	}
}

func (c *Client) GetReceiverInfo(recvKW [][]byte) *[]database.IKVElement {

	queriesGRPC := make([][]*pb.Query, c.NumServer)
	for i := 0; i < c.NumServer; i++ {
		queriesGRPC[i] = make([]*pb.Query, c.RateS*util.ARITY)
	}
	// Keep list of keywords and their according indices to find desired record
	// (and ignore dummy requests in non-auth case)
	queryKws := make([][]byte, c.RateS)

	// Client has to make fixed number of requests (rates*arity many),
	// if len(recvKW) < c.Rate S: generate dummy queries based on own idx
	var indices []uint32
	for i := 0; i < int(c.RateS); i++ {
		// add real queries
		if i < len(recvKW) {
			indices = c.Pps[database.Kw].GetIndices(recvKW[i])
			for j, idx := range indices {
				dpfKeys, _ := c.Dpfs[database.Kw].Query(int(idx))
				for k := 0; k < c.NumServer; k++ {
					queriesGRPC[k][i*util.ARITY+j] = &pb.Query{DpfKey: *dpfKeys[k]}
				}
			}
			queryKws[i] = recvKW[i]
		} else { // add dummy keywords and their indices
			for j := 0; j < util.ARITY; j++ {
				dpfKeys, _ := c.Dpfs[database.Kw].Query(int(c.Idx))
				for k := 0; k < c.NumServer; k++ {
					queriesGRPC[k][i*util.ARITY+j] = &pb.Query{DpfKey: *dpfKeys[k]}
				}
			}
			queryKws[i] = c.Id
		}
	}

	// Send all queries in parallel to servers
	ans_grpc := make([]*pb.Answers, c.NumServer)

	var wg sync.WaitGroup
	wg.Add(c.NumServer)
	for i := 0; i < int(c.NumServer); i++ {
		go makeQueriesWorker(c, &wg, i, &queriesGRPC, &ans_grpc, true)
	}
	wg.Wait()

	// Reconstruct DB records from answers
	var contactData []database.IKVElement
	for i, kw := range queryKws {

		// No need to check  Dummy Query in honest (not auth) setting
		if !c.Pps[database.Kw].Auth && bytes.Equal(c.Id, kw) {
			continue
		}
		for j := 0; j < util.ARITY; j++ {
			out, err := c.Dpfs[database.Kw].Reconstruct(
				[][]byte{ans_grpc[0].Answers[i*util.ARITY+j].Answer,
					ans_grpc[1].Answers[i*util.ARITY+j].Answer})
			if err != nil {
				log.Fatalf("failed to reconstruct answer")
			}
			if c.Pps[database.Kw].Auth {
				// Verify proof and remove proof from out
				out, err = c.Pps[database.Kw].VerifyRow(out)
				if err != nil {
					log.Fatalf("reject received pir answers, proof rejected")
				}
			}
			if bytes.Equal(out[:c.Pps[database.Kw].KeyLength], kw) {
				idx := util.ByteSliceToUint32(out[c.Pps[database.Kw].RecordLength-4:])
				contactData = append(contactData, c.Pps[database.Kw].RowToIKV(idx, out))
				break
			}
		}
	}
	return &contactData

}

func makeQueriesWorker(c *Client, wg *sync.WaitGroup, id int, queriesGRPC *[][]*pb.Query, ans_grpc *[]*pb.Answers, isSender bool) {
	defer wg.Done()
	var err error
	clientDeadline := time.Now().Add(util.TIMEOUT)
	ctx, cancel := context.WithTimeout(context.Background(), util.TIMEOUT)
	ctx, cancel = context.WithDeadline(ctx, clientDeadline)
	defer cancel()

	pb_in := &pb.Queries{Queries: (*queriesGRPC)[id]}
	if isSender {
		(*ans_grpc)[id], err = (*c.GrpcClients[id]).MakeKWQueries(ctx, pb_in)
	} else {
		(*ans_grpc)[id], err = (*c.GrpcClients[id]).MakeIQueries(ctx, pb_in)
	}

	if err != nil {
		log.Fatalf("could not get row: %v", err)
	}
	// BW cost is the same for each server, so we only need to measure it once and multiply by numserver
	if id == 0 {
		if isSender {
			c.BW["SendPIRUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["SendPIRDown"] += uint32(proto.Size((*ans_grpc)[id])) * uint32(c.NumServer)
		} else {
			c.BW["RecvPIRUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["RecvPIRDown"] += uint32(proto.Size((*ans_grpc)[id])) * uint32(c.NumServer)
		}
	}

}

func (c *Client) Notify(targets *[]database.IKVElement, isSender bool) {
	col := notify.CreateVectorIKV(targets, c.Pps[database.Idx].NRows)
	shares := notify.GenShares(col, c.NumServer, c.DpfSeed)

	var wg sync.WaitGroup
	wg.Add(c.NumServer)

	for i := 0; i < int(c.NumServer); i++ {
		go notifyWorker(c, &wg, i, &shares, isSender)
	}
	wg.Wait()
}

func notifyWorker(c *Client, wg *sync.WaitGroup, id int, shares *[][]byte, isSender bool) {
	defer wg.Done()

	clientDeadline := time.Now().Add(util.TIMEOUT)
	ctx, cancel := context.WithTimeout(context.Background(), util.TIMEOUT)
	ctx, cancel = context.WithDeadline(ctx, clientDeadline)
	defer cancel()

	pb_in := &pb.NotifyRequest{Idx: uint32(c.Idx), Vec: &pb.Vector{Val: (*shares)[id]}}
	pb_out, err := (*c.GrpcClients[id]).SetColumn(ctx, pb_in)
	if err != nil {
		log.Fatalf("could not write column: %v", err)
	}

	if id == 0 {
		if isSender {
			c.BW["SendNotifyUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["SendNotifyDown"] += uint32(proto.Size(pb_out)) * uint32(c.NumServer)
		} else {
			c.BW["RecvNotifyUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["RecvNotifyDown"] += uint32(proto.Size(pb_out)) * uint32(c.NumServer)
		}
	}

}

func (c *Client) GetNotified(isSender bool) []uint32 {
	shares := make([][]byte, c.NumServer)

	var wg sync.WaitGroup
	wg.Add(c.NumServer)
	for i := 0; i < int(c.NumServer); i++ {
		go getNotifiedWorker(c, &wg, i, &shares, isSender)
	}
	wg.Wait()
	row := notify.CombineShares(shares)
	//Get sender indices from row
	senders := notify.ReadVector(row)

	return senders
}
func getNotifiedWorker(c *Client, wg *sync.WaitGroup, id int, shares *[][]byte, isSender bool) {
	defer wg.Done()

	// Contact the server and print out its response.
	clientDeadline := time.Now().Add(util.TIMEOUT)
	ctx, cancel := context.WithTimeout(context.Background(), util.TIMEOUT)
	ctx, cancel = context.WithDeadline(ctx, clientDeadline)
	defer cancel()
	pb_in := &pb.Index{Idx: uint32(c.Idx)}
	sharedRow, err := (*c.GrpcClients[id]).GetRow(ctx, pb_in)
	if err != nil {
		log.Fatalf("could not get row: %v", err)
	}
	(*shares)[id] = sharedRow.Val
	if id == 0 {
		if isSender {
			c.BW["SendGetNotifiedUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["SendGetNotifiedDown"] += uint32(proto.Size(sharedRow)) * uint32(c.NumServer)
		} else {
			c.BW["RecvGetNotifiedUp"] += uint32(proto.Size(pb_in)) * uint32(c.NumServer)
			c.BW["RecvGetNotifiedDown"] += uint32(proto.Size(sharedRow)) * uint32(c.NumServer)
		}
	}
}

// Do index PIR for (all) senders based on retrieval rate
func (c *Client) GetSenders(senders []uint32) *[]database.IKVElement {
	// Client has to make fixed number of requests (rateR many)
	// generate dummy queries based on own idx
	for i := 0; i < int(c.RateR)-len(senders); i++ {
		senders = append(senders, c.Idx)
	}
	queriesGRPC := make([][]*pb.Query, c.NumServer)
	// generate all queries, reconstruct is just always the same here
	for _, sender := range senders {
		// generate dpf queries
		dpfKeys, _ := c.Dpfs[database.Idx].Query(int(sender))
		for k, key := range dpfKeys {
			var q_grpc = pb.Query{}
			q_grpc.DpfKey = (*key)[:]
			queriesGRPC[k] = append(queriesGRPC[k], &q_grpc)
		}
	}
	// Send all queries in parallel to servers
	ans_grpc := make([]*pb.Answers, c.NumServer)

	var wg sync.WaitGroup
	wg.Add(c.NumServer)
	for i := 0; i < int(c.NumServer); i++ {
		go makeQueriesWorker(c, &wg, i, &queriesGRPC, &ans_grpc, false)
	}
	wg.Wait()

	var senderData []database.IKVElement
	for i, senderIdx := range senders {
		if !c.Pps[database.Idx].Auth && senderIdx != c.Idx {
			out, err := c.Dpfs[database.Idx].Reconstruct(
				[][]byte{ans_grpc[0].Answers[i].Answer,
					ans_grpc[1].Answers[i].Answer})
			if err != nil {
				log.Fatalf("failed to reconstruct answer")
			}
			senderData = append(senderData, c.Pps[database.Idx].RowToIKV(senderIdx, out))
		}
		if c.Pps[database.Idx].Auth {
			// all queries in auth case have to be checked to ensure server learns nothing
			out, err := c.Dpfs[database.Idx].Reconstruct(
				[][]byte{ans_grpc[0].Answers[i].Answer,
					ans_grpc[1].Answers[i].Answer})
			if err != nil {
				log.Fatalf("failed to reconstruct answer")
			}
			// Verify proof
			data, err := c.Pps[database.Idx].VerifyRow(out)
			if err != nil {
				log.Fatalf("reject received pir answers, proof rejected")
			}
			// if it was not a dummy query, add the info to the sender list
			if senderIdx != c.Idx {
				senderData = append(senderData, c.Pps[database.Idx].RowToIKV(senderIdx, data))
			}
		}
	}
	return &senderData
}
