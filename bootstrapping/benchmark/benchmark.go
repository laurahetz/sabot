package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"runtime"
	bs "sabot/bootstrapping"
	"sabot/lib/database"
	"time"
)

const (
	defaultPath = "/app/benchmarks/configs.json"
	defaultOut  = "/app/benchmarks/results.csv"
	defaultReps = 20
	INPUT_SEED  = 42
)

var (
	pathRead  = flag.String("path", defaultPath, "path for reading benchmark configs.")
	pathWrite = flag.String("out", defaultOut, "path for writing benchmark results.")
)

/*
read in JSON file, that contains multiple configs
each config is one benchmark test for which a new client is generated.
the client runs the SetupExperiment call to set up the servers.
C and S run the experiment
*/
func main() {
	flag.Parse()

	rConfig := bs.ReadBenchConfigs(*pathRead)
	file, err := os.Create(*pathWrite)
	if err != nil {
		log.Fatal("error creating file", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write column headers
	writer.Write(bs.Headers)

	for _, config := range rConfig.Configs {
		var start time.Time

		// SETUP:  Init Client and Server
		c := bs.InitClient(&config, &bs.ServerInfo{Addr: []string{rConfig.Addr1, rConfig.Addr2}})
		// For Benchmarking: Get random keywords (that are included in the database)
		// and the client's kw from server
		recvKWs := make([][]byte, c.RateS)
		for i, contact := range *c.Contacts {
			recvKWs[i] = contact.Key
		}
		for i := 0; i < int(config.Repetitions); i++ {
			//	Sender Retrieval
			// Run KW PIR to get contact info of receivers
			start = time.Now()
			receivers := c.GetReceiverInfo(recvKWs)
			c.RT["SendPIR"] += time.Since(start)

			//	Sender Notification
			start = time.Now()
			c.Notify(receivers, true)
			c.RT["SendNotify"] += time.Since(start)

			// Receiver GetNotificaion
			start = time.Now()
			senderIndices := c.GetNotified(false)
			c.RT["RecvGetNotified"] += time.Since(start)

			// Receiver Retrieval
			start = time.Now()
			senders := c.GetSenders(senderIndices)
			c.RT["RecvPIR"] += time.Since(start)

			// Receiver Notification
			start = time.Now()
			c.Notify(senders, false)
			c.RT["RecvNotify"] += time.Since(start)

			// Sender getNotified
			start = time.Now()
			// client should match based on this info with whom they have now exchanged infos
			c.GetNotified(true)
			c.RT["SendGetNotified"] += time.Since(start)

			runtime.GC()

			// get output as string for result file
			out := bs.GetOutputString(c.Experiment, *c.Pps[database.Idx], i)
			writer.Write(out)
			writer.Flush()

			// Reset exp, BW and RT summary
			c.ResetBenchVars()
		}
		for _, conn := range c.ServerInfo.Conns {
			conn.Close()
		}

		writer.Flush()
	}
	file.Close()
}
