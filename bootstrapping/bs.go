package bootstrapping

import (
	"encoding/json"
	"log"
	"os"
	"sabot/lib/database"
	"strconv"
	"time"
)

// Experiment Config
type Config struct {
	Idx         uint32
	Dbfile      string
	RateR       uint32
	RateS       uint32
	MultiClient bool
	NumThreads  uint32
	ResetServer bool
	Repetitions uint32
	DBType      uint32 // 0: 2 DBs
}

// Experiment Suite
// used to read experiment configs from file
type DriverConfig struct {
	Addr1   string
	Addr2   string
	Configs []Config
}

type Experiment struct {
	*Config
	BW map[string]uint32
	RT map[string]time.Duration
}

func NewExperiment(config *Config) *Experiment {
	exp := Experiment{}
	exp.Config = config
	exp.ResetBenchVars()

	return &exp
}

func (exp *Experiment) ResetBenchVars() {
	exp.BW = map[string]uint32{
		"SendNotifyUp":        0,
		"SendNotifyDown":      0,
		"RecvNotifyUp":        0,
		"RecvNotifyDown":      0,
		"SendGetNotifiedUp":   0,
		"SendGetNotifiedDown": 0,
		"RecvGetNotifiedUp":   0,
		"RecvGetNotifiedDown": 0,
		"SendPIRUp":           0,
		"SendPIRDown":         0,
		"RecvPIRUp":           0,
		"RecvPIRDown":         0,
	}
	exp.RT = map[string]time.Duration{
		"SendNotify":      0,
		"RecvNotify":      0,
		"SendGetNotified": 0,
		"RecvGetNotified": 0,
		"SendPIR":         0,
		"RecvPIR":         0,
	}
}

// Input Config, DBParams and number of repetition
func GetOutputString(exp *Experiment, pp database.DBParams, i int) []string {
	var out []string

	for _, key := range Headers {
		if key == "db_type" {
			dbT := database.DBType(exp.DBType)
			log.Println("\ndb_type:", key, ":", dbT.String())
			out = append(out, dbT.String())
		} else if key == "db_size" {
			log.Println("\ndb_size:", key, ":", pp.NRows)
			out = append(out, strconv.Itoa(int(pp.NRows)))
		} else if key == "key_length" {
			log.Println("key_length:", strconv.Itoa(int(pp.KeyLength)))
			out = append(out, strconv.Itoa(int(pp.KeyLength)))
		} else if key == "value_length" {
			log.Println("value_length:", strconv.Itoa(int(pp.ValueLength)))
			out = append(out, strconv.Itoa(int(pp.ValueLength)))
		} else if key == "malicious" {
			log.Println("malicious:", pp.Auth)
			out = append(out, strconv.FormatBool(pp.Auth))
		} else if key == "rate" {
			log.Println("rate:", strconv.Itoa(int(exp.RateR)))
			out = append(out, strconv.Itoa(int(exp.RateR)))
		} else if key == "multi_client" {
			log.Println("multi_client:", strconv.FormatBool(exp.MultiClient))
			out = append(out, strconv.FormatBool(exp.MultiClient))
		} else if key == "num_threads" {
			log.Println("num_threads:", strconv.Itoa(int(exp.NumThreads)))
			out = append(out, strconv.Itoa(int(exp.NumThreads)))
		} else if key == "repetition" {
			log.Println("rep:", strconv.Itoa(i))
			out = append(out, strconv.Itoa(i))
		} else if key[:3] == "BW_" {
			log.Println("BW:", key, ":", strconv.Itoa(int(exp.BW[key[3:]])))
			out = append(out, strconv.Itoa(int(exp.BW[key[3:]])))
		} else if key[:3] == "RT_" {
			log.Println("RT:", key, ":", strconv.Itoa(int(exp.RT[key[3:]].Microseconds())))
			out = append(out, strconv.Itoa(int(exp.RT[key[3:]].Microseconds())))
		}
	}
	return out
}

// read in JSON file, that contains multiple configs
// each config is one benchmark test for which a new client is generated
func ReadBenchConfigs(path string) *DriverConfig {

	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var rConfig DriverConfig
	err = json.Unmarshal(content, &rConfig)

	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	return &rConfig

}

// define column headers
var Headers = []string{
	"db_type",
	"db_size",
	"key_length",
	"value_length",
	"malicious",
	"rate",
	"multi_client",
	"num_threads",
	"repetition",
	"BW_SendPIRUp",
	"BW_SendPIRDown",
	"BW_SendNotifyUp",
	"BW_SendNotifyDown",
	"BW_RecvGetNotifiedUp",
	"BW_RecvGetNotifiedDown",
	"BW_RecvPIRUp",
	"BW_RecvPIRDown",
	"BW_RecvNotifyUp",
	"BW_RecvNotifyDown",
	"BW_SendGetNotifiedUp",
	"BW_SendGetNotifiedDown",
	"RT_SendPIR",
	"RT_SendNotify",
	"RT_RecvGetNotified",
	"RT_RecvPIR",
	"RT_RecvNotify",
	"RT_SendGetNotified",
}
