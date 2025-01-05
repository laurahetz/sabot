package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	bs "sabot/bootstrapping"
	"sabot/lib/database"
	"sabot/lib/notify"
	"sabot/lib/util"
	pb "sabot/proto/bootstrapping"

	"google.golang.org/grpc"
)

const (
	localTestPrefix  = ""
	localDebugPrefix = ""
)

var (
	port = flag.Int("port", 50051, "server port")
)

type gRPCServer struct {
	pb.UnimplementedBootstrappingServer
	*bs.Server
}

func (s *gRPCServer) SetColumn(ctx context.Context, in *pb.NotifyRequest) (*pb.Ack, error) {
	if s.Server == nil {
		return nil, errors.New("server not initialized")
	}
	s.Server.SetColumn(in)
	return &pb.Ack{Ok: true}, nil
}

func (s *gRPCServer) GetRow(ctx context.Context, in *pb.Index) (*pb.Vector, error) {
	if s.Server == nil {
		return nil, errors.New("server not initialized")
	}
	out := s.Server.GetRow(in.Idx)
	return &pb.Vector{Val: out}, nil
}

func (s *gRPCServer) MakeIQueries(ctx context.Context, in *pb.Queries) (*pb.Answers, error) {
	if s.Server == nil {
		return nil, errors.New("server not initialized")
	}
	out, err := s.Server.AnswerIQueries(in)

	return &pb.Answers{Answers: out}, err
}

func (s *gRPCServer) MakeKWQueries(ctx context.Context, in *pb.Queries) (*pb.Answers, error) {
	if s.Server == nil {
		return nil, errors.New("server not initialized")
	}
	out, err := s.Server.AnswerKWQueries(in)

	return &pb.Answers{Answers: out}, err
}

/*
Server obtains config for experiment to run from the client,
config includes which DB file to read in and use and other parameters

Server returns the protocol's parameters to the client
*/
func (s *gRPCServer) SetupExperiment(ctx context.Context, in *pb.Config) (*pb.ParamResp, error) {
	if in.ResetServer {
		s.Server = &bs.Server{}
		// Set DB Type and read DB(s) from disk
		s.ContactDB = &database.ContactDB{DBType: database.DBType(util.ByteSliceToUint32(in.DbType))}
		log.Println("reset server: init from File(s):", in.Dbfile, "dbtype: ", s.ContactDB.DBType)
		s.ContactDB.FromDisk(localTestPrefix + in.Dbfile)

		// Set size of notification matrix to size of index database
		s.NotifyMatrix = notify.NewMatrix(int(s.DBs[database.Idx].Db.NumRows))

	} else if s.Server == nil {
		log.Fatalln("server not initialized!")
	}
	// Set all other server config parameters

	s.MultiClient = in.MultiClient
	s.NumThreads = int(in.NumThreads)

	// For Benchmarking Purposes
	// Return the clients KW and some existing target keywords to the client
	cKW, targets := s.Server.GetClientSetupValues(in.CIdx, in.NumTargets)

	params := make([]*pb.Params, len(s.DBs))
	for i, db := range s.DBs {
		params[i] = &pb.Params{
			Nrows:       db.Pp.NRows,
			Auth:        db.Pp.Auth,
			Seed:        db.Pp.Seed,
			SegLen:      db.Pp.SegmentLength,
			SegLenMask:  db.Pp.SegmentLengthMask,
			SegCount:    db.Pp.SegmentCount,
			SegCountLen: db.Pp.SegmentCountLength,
			KeyLen:      db.Pp.KeyLength,
			ValLen:      db.Pp.ValueLength,
			ProofLen:    db.Pp.ProofLen,
			Root:        db.Pp.Root,
			RecLength:   db.Pp.RecordLength,
		}
	}

	return &pb.ParamResp{
		CKW:     cKW,
		Params:  params,
		Targets: targets,
	}, nil
}

func main() {
	flag.Parse()
	// Open port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// false = server running this function
	creds, err := util.LoadTLSCred(localDebugPrefix+util.CERT_S_PATH_PRE, localDebugPrefix+util.CERT_CA_PATH, false)
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.MaxRecvMsgSize(util.MAX_MSG_SIZE),
		grpc.MaxSendMsgSize(util.MAX_MSG_SIZE),
	)
	grpcServer := &gRPCServer{}
	pb.RegisterBootstrappingServer(s, grpcServer)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
