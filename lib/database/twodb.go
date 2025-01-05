package database

import (
	"log"
	"sabot/lib/util"
)

func SetupTwoDBs(tempDb *[]IKVElement, pp *DBParams, size uint32) (dbs []*Database, err error) {
	dbs = make([]*Database, 2)
	dbs[Idx] = &Database{}
	dbs[Kw] = &Database{}

	tempKWDB := make([][]byte, len(*tempDb))
	tempIDB := make([][]byte, size)

	dbs[Idx].Pp = *pp
	// tempDB contains empty slots, but Index DB only holds non-empty records
	dbs[Idx].Pp.NRows = size
	dbs[Idx].Pp.RecordLength = pp.KeyLength + pp.ValueLength

	dbs[Kw].Pp = *pp
	// KW DB also holds non-empty records
	dbs[Kw].Pp.NRows = uint32(len(tempKWDB))
	dbs[Kw].Pp.RecordLength = pp.KeyLength + pp.ValueLength + 4 //including 4 bytes for ipir index

	var ctr uint32 = 0
	for i, ikv := range *tempDb {
		if len(ikv.Value) == 0 {
			tempKWDB[i] = make([]byte, dbs[Kw].Pp.RecordLength)
			ctr++
		} else {
			// add i-DB index of each value to KW-DB
			tempKWDB[i] = append(ikv.Key, append(ikv.Value, util.Uint32ToByteSlice(uint32(i)-ctr)...)...)
			tempIDB[uint32(i)-ctr] = append(ikv.Key, ikv.Value...)
		}
	}

	// Add merkle proofs in auth case, updates tempDB directly
	if pp.Auth {
		dbs[Idx].Pp.CreateMerkle(&tempIDB)
		dbs[Kw].Pp.CreateMerkle(&tempKWDB)
	} else {
		dbs[Idx].Pp.Root = []byte{}
		dbs[Idx].Pp.ProofLen = 0
		dbs[Kw].Pp.Root = []byte{}
		dbs[Kw].Pp.ProofLen = 0
	}

	dbs[Idx].Db, err = StaticDBFromRows(tempIDB)
	if err != nil {
		log.Fatalln("error creating database from rows")
	}
	dbs[Kw].Db, err = StaticDBFromRows(tempKWDB)
	if err != nil {
		log.Fatalln("error creating database from rows")
	}
	return
}
