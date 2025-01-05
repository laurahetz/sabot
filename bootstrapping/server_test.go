package bootstrapping

import (
	"bytes"
	"log"
	"sabot/lib/database"
	"sabot/lib/notify"
	"sabot/lib/util"
	"testing"
)

func TestServer(t *testing.T) {

	keylen := util.KEY_LENGTH
	valuelen := util.VAL_LENGTH
	auth := false
	var numInputs uint32 = 100
	dbtype := database.TwoDB

	s := Server{
		ContactDB:   &database.ContactDB{DBType: dbtype},
		MultiClient: false,
		NumThreads:  1,
	}

	// Get db test values
	inputs := database.GetTestData(numInputs, uint(keylen), uint(valuelen), 42)

	s.ContactDB.Setup(inputs, auth)
	s.NotifyMatrix = notify.NewMatrix(s.DBs[database.Idx].Db.NumRows)

	numTargets := 10
	// Test get Client Value functionality for all elements
	for i := 0; i < s.DBs[database.Idx].Db.NumRows; i++ {
		cKW, targets := s.GetClientSetupValues(uint32(i), uint32(numTargets))
		for j := 0; j < numTargets; j++ {
			if bytes.Equal(cKW, targets[j*keylen:(j+1)*keylen]) {
				log.Fatalln("targets contains cKW, but this should be excluded")
			}
		}
	}
}
