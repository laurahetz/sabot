package main

import (
	"flag"
	"log"
	"sabot/lib/database"
	"strconv"
	"time"
)

const (
	defaultSizeExp = 10
	arity          = 3  // BFF setup parameter, num hash fcts
	kvSeed         = 42 // To generate test data
	defaultKeyLen  = 32 // 256 bit
	defaultValLen  = 32 // 256 bit
	defaultAuth    = false
	prefix         = "/app/db/db_" // TO CHANGE if running this locally
)

var (
	sizeExp = flag.Uint("sizeExp", defaultSizeExp, "size of database (2^x)")
	keyLen  = flag.Uint("keyLen", defaultKeyLen, "size of client identifier in byte")
	valLen  = flag.Uint("valLen", defaultValLen, "size of client contact info in byte")
	auth    = flag.Bool("auth", defaultAuth, "use authenticated mode")
	path    = flag.String("path", "", "path for storing db in file. Default: db_sizeExp_keyLen_valLen_auth.db")
	dbtype  = flag.Uint("dbtype", uint(database.TwoDB.EnumIndex()), "Type of database to use: 0 (TwoDB, default)")
)

func main() {
	flag.Parse()

	numClients := uint32(1 << *sizeExp)
	// generate Test data
	input := database.GetTestData(numClients, *keyLen, *valLen, kvSeed)

	start := time.Now()

	cdb := database.ContactDB{DBType: database.DBType(*dbtype)}
	cdb.Setup(input, *auth)
	t := time.Since(start)
	log.Println("RT DBGen for N_client = 2^", *sizeExp, ", keyLen=", *keyLen, ", valLen=", *valLen, ", APIR=", *auth, ":", t)

	if *path == "" {
		*path = prefix + strconv.Itoa(int(*sizeExp)) + "_" + strconv.Itoa(int(*keyLen)) + "_" +
			strconv.Itoa(int(*valLen)) + "_" + strconv.FormatBool(*auth)
	}
	cdb.ToDisk(*path)
}
