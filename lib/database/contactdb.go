package database

import (
	"log"
	"sabot/lib/util"
)

// enum item to specify how the database is set up
type DBType int
type QueryType int

const (
	TwoDB DBType = iota // EnumIndex = 0
)

const (
	Idx QueryType = iota // EnumIndex = 0
	Kw                   // EnumIndex = 1
)

const (
	IPIR_EXT  = ".ipir"
	KWPIR_EXT = ".kwpir"
)

func (d DBType) String() string {
	return [...]string{"TwoDB"}[d]
}

func (d DBType) EnumIndex() int {
	return int(d)
}

func (d QueryType) String() string {
	return [...]string{"Kw", "Idx"}[d]
}

func (d QueryType) EnumIndex() int {
	return int(d)
}

type ContactDB struct {
	DBType
	DBs []*Database
}

func (cdb *ContactDB) Setup(inputs []KVElement, auth bool) {
	// do BFF Setup
	pp, tempDB, err := SetupBinaryFuse(inputs, util.ARITY)
	if err != nil {
		log.Fatalln("BFF setup failed")
	}
	pp.Auth = auth

	if cdb.DBType == TwoDB {
		cdb.DBs, err = SetupTwoDBs(tempDB, pp, uint32(len(inputs)))
		if err != nil {
			log.Fatalln("error creating databases")
		}
	} else {
		log.Fatalln("other DBTypes not supported")
	}
}

func (cdb *ContactDB) ToDisk(path string) {
	if cdb.DBType == TwoDB {
		ContactDBToFile(path+IPIR_EXT, cdb.DBs[Idx])
		ContactDBToFile(path+KWPIR_EXT, cdb.DBs[Kw])
	} else {
		log.Fatalln("other DBTypes not supported")
	}
}

func (cdb *ContactDB) FromDisk(path string) {

	if cdb.DBType == TwoDB {
		cdb.DBs = make([]*Database, 2)
		cdb.DBs[Idx] = ContactDBFromFile(path + IPIR_EXT)
		cdb.DBs[Kw] = ContactDBFromFile(path + KWPIR_EXT)
	} else {
		log.Fatalln("other DBTypes not supported")
	}
}
