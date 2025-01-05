package pir

import (
	"math/rand"
	"sabot/lib/database"
	"sabot/lib/util"
)

var masterKey util.PRGKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 'A', 'B', 'C', 'D', 'E', 'F'}
var randReader *rand.Rand = rand.New(util.NewBufPRG(util.NewPRG(&masterKey)))

func RandSource() *rand.Rand {
	return randReader
}

func MakeRows(src *rand.Rand, nRows, rowLen int) [][]byte {
	db := make([][]byte, nRows)
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
		db[i][0] = byte(i % 256)
		db[i][1] = 'A' + byte(i%256)
	}
	return db
}

func MakeDB(nRows int, rowLen int) *database.StaticDB {
	rows := MakeRows(RandSource(), nRows, rowLen)
	db, err := database.StaticDBFromRows(rows)
	if err == nil {
		return db
	}
	return nil
}

func MakeKeys(src *rand.Rand, nRows int) []uint32 {
	keys := make([]uint32, nRows)
	for i := range keys {
		keys[i] = uint32(src.Int31())
	}
	return keys
}

func MakeKeysRows(numRows, rowLen int) ([]uint32, [][]byte) {
	return MakeKeys(RandSource(), numRows), MakeRows(RandSource(), numRows, rowLen)
}
