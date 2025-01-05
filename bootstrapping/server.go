package bootstrapping

import (
	"log"
	"math/rand"
	"sabot/lib/database"
	"sabot/lib/notify"
	"sabot/lib/pir"
	"sabot/lib/util"
	pb "sabot/proto/bootstrapping"
	"sync"

	"github.com/dkales/dpf-go/dpf"
)

type Server struct {
	*notify.NotifyMatrix
	*database.ContactDB
	MultiClient bool
	NumThreads  int
}

func answerQueriesWorker(db *database.Database, id int, jobs <-chan *pb.Queries, wg *sync.WaitGroup, answers *[]*pb.Answer) {
	for q := range jobs {
		var resp *pir.DPFQueryResp
		var err error
		for i, query := range q.Queries {
			resp, err = pir.Process(db.Db, (*dpf.DPFkey)(&query.DpfKey))
			if err != nil {
				log.Fatal("error processing query")
			}
			(*answers)[i] = &pb.Answer{Answer: resp.Answer}
		}
		log.Printf("Thread %d has done some work!\n", id)
		wg.Done()
	}

}

func (s *Server) AnswerQueries(in *pb.Queries, queryType database.QueryType) ([]*pb.Answer, error) {
	var wg sync.WaitGroup
	// one job per client
	// if singleClient experiment numJubs = 1, else its |Index-PIR DB|
	var numJobs int
	if !s.MultiClient {
		numJobs = 1
	} else {
		// This is independent of the queryType
		numJobs = s.DBs[database.Idx].Db.NumRows
	}
	wg.Add(int(numJobs))
	jobs := make(chan *pb.Queries, numJobs)
	answers := make([]*pb.Answer, len(in.Queries))

	for w := 0; w < s.NumThreads; w++ {
		go answerQueriesWorker(s.DBs[queryType], w, jobs, &wg, &answers)
	}
	for j := 0; j < int(numJobs); j++ {
		jobs <- in
	}
	close(jobs)
	wg.Wait()

	return answers, nil
}

func (s *Server) AnswerIQueries(in *pb.Queries) ([]*pb.Answer, error) {
	return s.AnswerQueries(in, database.Idx)
}

func (s *Server) AnswerKWQueries(in *pb.Queries) ([]*pb.Answer, error) {
	return s.AnswerQueries(in, database.Kw)
}

func (s *Server) SetColumn(in *pb.NotifyRequest) {

	var wg sync.WaitGroup
	var numJobs int
	if !s.MultiClient {
		numJobs = 1
	} else {
		numJobs = s.NotifyMatrix.NumRows
	}
	wg.Add(int(numJobs))
	jobs := make(chan *pb.NotifyRequest, numJobs)

	for w := 0; w < s.NumThreads; w++ {
		go setColWorker(s, w, jobs, &wg)
	}
	for j := 0; j < int(numJobs); j++ {
		jobs <- in
	}
	close(jobs)
	wg.Wait()
}
func setColWorker(s *Server, id int, jobs <-chan *pb.NotifyRequest, wg *sync.WaitGroup) {
	for i := range jobs {
		// Add column to matrix
		s.NotifyMatrix.SetColumn(int(i.Idx), i.Vec.Val)
		log.Printf("Thread %d has done some work!\n", id)
		wg.Done()
	}
}

func (s *Server) GetRow(idx uint32) []byte {
	var wg sync.WaitGroup
	var numJobs int
	if !s.MultiClient {
		numJobs = 1

	} else {
		numJobs = s.NotifyMatrix.NumRows
	}
	wg.Add(int(numJobs))
	jobs := make(chan uint32, numJobs)
	rows := make([][]byte, numJobs)

	for w := 0; w < s.NumThreads; w++ {
		go getRowWorker(s, w, jobs, &wg, &rows)
	}
	for j := 0; j < int(numJobs); j++ {
		jobs <- uint32(j)
	}
	close(jobs)
	wg.Wait()

	return rows[idx]
}

func getRowWorker(s *Server, id int, jobs <-chan uint32, wg *sync.WaitGroup, rows *[][]byte) {
	for i := range jobs {
		(*rows)[i] = s.NotifyMatrix.GetRow(i)
		if (*rows)[i] == nil {
			log.Fatal("row index out of range")
		}
		log.Printf("Thread %d has done some work!\n", id)
		wg.Done()
	}
}

func (s *Server) GetClientSetupValues(cid uint32, numTargets uint32) ([]byte, []byte) {

	ckw := s.DBs[database.Idx].Db.Row(int(cid))[:util.KEY_LENGTH]

	targets := make([]byte, numTargets*util.KEY_LENGTH)
	r := rand.New(rand.NewSource(util.INPUT_SEED))
	targetsIdx := database.RandTargetsExcept(r, int(numTargets), int(s.DBs[database.Idx].Db.NumRows-1), 0, cid)
	for i, idx := range targetsIdx {
		copy(targets[i*int(util.KEY_LENGTH):(i+1)*int(util.KEY_LENGTH)], s.DBs[database.Idx].Db.Row(int(idx))[:util.KEY_LENGTH])
	}

	return ckw, targets
}
