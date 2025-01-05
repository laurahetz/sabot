package database

import (
	"bytes"
	"reflect"

	"sabot/lib/util"
	"testing"
)

func TestUint32ToByte(t *testing.T) {
	var i uint32 = 2

	i_byte := util.Uint32ToByteSlice(i)
	i_uint32 := util.ByteSliceToUint32(i_byte)
	if i != i_uint32 {
		t.Fatalf("transform not successful")
	}
}

func TestBFFSetup(t *testing.T) {
	keylen := util.KEY_LENGTH
	valuelen := util.VAL_LENGTH

	elements := GetTestData(10000, uint(keylen), uint(valuelen), 42)

	pp, f, err := SetupBinaryFuse(elements, 3)

	if err != nil {
		t.Fatal("BFF Setup failed:", err)
	}

	for _, element := range elements {
		indices := pp.GetIndices(element.Key)
		foundIdx := -1
		for _, i := range indices {
			if len((*f)[i].Value) > 0 {
				if reflect.DeepEqual((*f)[i].Key, element.Key) {
					foundIdx = int(i)
					break
				}
			}
		}
		if foundIdx == -1 {
			t.Fatalf("Key not in BFF")
		}
		// Check if value matches
		if !reflect.DeepEqual((*f)[foundIdx].Value, element.Value) {
			t.Fatalf("Value for key is not correct")
		}
		// Check if value matches
		if !reflect.DeepEqual((*f)[foundIdx].Idx, uint32(foundIdx)) {
			t.Fatalf("Value for idx is not correct")
		}
	}
}

func TestDBSetupAndGet(t *testing.T) {
	keylen := util.KEY_LENGTH
	valuelen := util.VAL_LENGTH
	numElements := 10000
	auth := false
	dbtype := TwoDB

	input := GetTestData(uint32(numElements), uint(keylen), uint(valuelen), 42)

	cdb := ContactDB{DBType: dbtype}
	cdb.Setup(input, auth)

	for _, element := range input {
		indices := cdb.DBs[Idx].Pp.GetIndices(element.Key)

		contains, ikv := cdb.DBs[Kw].Get(element.Key)
		if !contains {
			t.Fatal("Failed to find element in DB")
		}

		idxMatch := false
		idx := 0
		for _, i := range indices {
			if reflect.DeepEqual(ikv.Idx, i) {
				idxMatch = true
				idx = int(i)
				break
			}
		}
		if !idxMatch {
			t.Fatalf("Index incorrect")
		}

		if !reflect.DeepEqual(ikv.Key, element.Key) {
			t.Fatal("Key incorrect")

		}
		if !reflect.DeepEqual(ikv.Value, element.Value) {
			t.Fatal("Value incorrect")
		}

		kwrow := cdb.DBs[Kw].Db.Row(idx)
		idb_idx := util.ByteSliceToUint32(kwrow[len(kwrow)-4:])
		irow := cdb.DBs[Idx].Db.Row(int(idb_idx))

		if !bytes.Equal(kwrow[:keylen], irow[:keylen]) {
			t.Fatal("keywords in dbs not equal")
		}
		if !bytes.Equal(kwrow[keylen:valuelen+keylen], irow[keylen:valuelen+keylen]) {
			t.Fatal("values in dbs not equal")
		}
	}
}

func TestDBSetupAndGetAuth(t *testing.T) {
	keylen := util.KEY_LENGTH
	valuelen := util.VAL_LENGTH
	numElements := 10000
	auth := true
	dbtype := TwoDB

	input := GetTestData(uint32(numElements), uint(keylen), uint(valuelen), 42)

	cdb := ContactDB{DBType: dbtype}
	cdb.Setup(input, auth)

	for _, element := range input {
		indices := cdb.DBs[Kw].Pp.GetIndices(element.Key)
		contains, ikv := cdb.DBs[Kw].Get(element.Key)
		if !contains {
			t.Fatal("Failed to find element in DB")
		}
		// check that the  BFF setup has assigned one of the values in indices as the primary one
		idxMatch := false
		idx := 0
		for _, i := range indices {
			if reflect.DeepEqual(ikv.Idx, i) {
				idxMatch = true
				idx = int(i)
				break
			}
		}
		if !idxMatch {
			t.Fatalf("Index incorrect")
		}

		if !reflect.DeepEqual(ikv.Key, element.Key) {
			t.Fatal("Key incorrect")

		}
		if !reflect.DeepEqual(ikv.Value, element.Value) {
			t.Fatal("Value incorrect")
		}

		kwrow := cdb.DBs[Kw].Db.Row(idx)
		start := int(cdb.DBs[Kw].Pp.KeyLength + cdb.DBs[Kw].Pp.ValueLength)
		idb_idx := util.ByteSliceToUint32(kwrow[start : start+4])
		irow := cdb.DBs[Idx].Db.Row(int(idb_idx))

		if !bytes.Equal(kwrow[:keylen], irow[:keylen]) {
			t.Fatal("keywords in dbs not equal")
		}
		if !bytes.Equal(kwrow[keylen:valuelen+keylen], irow[keylen:valuelen+keylen]) {
			t.Fatal("values in dbs not equal")
		}
	}
}
