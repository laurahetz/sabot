package database

import (
	"bytes"
	"sabot/lib/util"
	"testing"
)

func TestContactDBFile(t *testing.T) {
	numKeys := 262144 //20000
	keyLength := util.KEY_LENGTH
	valLength := util.VAL_LENGTH
	seed := 42
	dbtype_b := util.Uint32ToByteSlice(0)
	dbType := DBType(util.ByteSliceToUint32(dbtype_b))
	//dbType := TwoDB
	auth := false
	input := GetTestData(uint32(numKeys), uint(keyLength), uint(valLength), int64(seed))

	cDB := ContactDB{DBType: DBType(dbType)}

	cDB.Setup(input, auth)

	path := "test"
	ContactDBToFile(path+IPIR_EXT, cDB.DBs[Idx])
	ContactDBToFile(path+KWPIR_EXT, cDB.DBs[Kw])

	idb := ContactDBFromFile(path + IPIR_EXT)
	kwdb := ContactDBFromFile(path + KWPIR_EXT)

	if !EqualPublicParams(&cDB.DBs[Idx].Pp, &idb.Pp) {
		t.Fatal("cDB.DBs[Idx] Params written and read are not the same")
	}
	if cDB.DBs[Idx].Db.NumRows != idb.Db.NumRows {
		t.Fatal("cDB.DBs[Idx].Db.NumRows written and read are not the same")
	}
	if cDB.DBs[Idx].Db.RowLen != idb.Db.RowLen {
		t.Fatal("cDB.DBs[Idx].Db.RowLen written and read are not the same")
	}
	if !bytes.Equal(cDB.DBs[Idx].Db.FlatDb, idb.Db.FlatDb) {
		t.Fatal("cDB.DBs[Idx].Db written and read are not the same")
	}

	if !EqualPublicParams(&cDB.DBs[Kw].Pp, &kwdb.Pp) {
		t.Fatal("KW DB Params written and read are not the same")
	}
	if cDB.DBs[Kw].Db.NumRows != kwdb.Db.NumRows {
		t.Fatal("cDB.DBs[Kw].Db.NumRows written and read are not the same")
	}
	if cDB.DBs[Kw].Db.RowLen != kwdb.Db.RowLen {
		t.Fatal("cDB.DBs[Kw].Db.RowLen written and read are not the same")
	}
	if !bytes.Equal(cDB.DBs[Kw].Db.FlatDb, kwdb.Db.FlatDb) {
		t.Fatal("cDB.DBs[Kw].Db written and read are not the same")
	}
}

func TestContactDBToDisk(t *testing.T) {
	numKeys := 262144 //20000
	keyLength := util.KEY_LENGTH
	valLength := util.VAL_LENGTH
	seed := 42
	dbType := TwoDB
	//auth := false
	path := "test"

	input := GetTestData(uint32(numKeys), uint(keyLength), uint(valLength), int64(seed))

	for _, auth := range []bool{false, true} {

		cDB := ContactDB{DBType: dbType}
		cDB.Setup(input, auth)
		cDB.ToDisk(path)

		cDB2 := ContactDB{DBType: dbType}
		cDB2.FromDisk(path)

		if !EqualPublicParams(&cDB.DBs[Idx].Pp, &cDB2.DBs[Idx].Pp) {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Idx] Params written and read are not the same")
		}
		if cDB.DBs[Idx].Db.NumRows != cDB2.DBs[Idx].Db.NumRows {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Idx].Db.NumRows written and read are not the same")
		}
		if cDB.DBs[Idx].Db.RowLen != cDB2.DBs[Idx].Db.RowLen {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Idx].Db.RowLen written and read are not the same")
		}
		if !bytes.Equal(cDB.DBs[Idx].Db.FlatDb, cDB2.DBs[Idx].Db.FlatDb) {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Idx].Db written and read are not the same")
		}

		if !EqualPublicParams(&cDB.DBs[Kw].Pp, &cDB2.DBs[Kw].Pp) {
			t.Fatal("auth: ", auth, "\tKW DB Params written and read are not the same")
		}
		if cDB.DBs[Kw].Db.NumRows != cDB2.DBs[Kw].Db.NumRows {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Kw].Db.NumRows written and read are not the same")
		}
		if cDB.DBs[Kw].Db.RowLen != cDB2.DBs[Kw].Db.RowLen {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Kw].Db.RowLen written and read are not the same")
		}
		if !bytes.Equal(cDB.DBs[Kw].Db.FlatDb, cDB2.DBs[Kw].Db.FlatDb) {
			t.Fatal("auth: ", auth, "\tcDB.DBs[Kw].Db written and read are not the same")
		}

	}
}
