package database

import (
	"errors"

	"github.com/lukechampine/fastxor"
)

type StaticDB struct {
	NumRows int // in notification case this is equal to num columns
	RowLen  int
	FlatDb  []byte
}

func (db *StaticDB) Slice(start, end int) []byte {
	return db.FlatDb[start*db.RowLen : end*db.RowLen]
}

func (db *StaticDB) Row(i int) []byte {
	if i >= db.NumRows {
		return nil
	}
	return db.Slice(i, i+1)
}

// input array of byte arrays instead of array of type Row
func StaticDBFromRows(data [][]byte) (*StaticDB, error) {
	var err error
	if len(data) < 1 {
		return &StaticDB{0, 0, nil}, nil
	}

	rowLen := len(data[0])
	flatDb := make([]byte, rowLen*len(data))

	for i, v := range data {
		if len(v) != rowLen {
			//fmt.Printf("Got row[%v] %v %v\n", i, len(v), rowLen)
			err = errors.New("Database rows must all be of the same length")
			break
		}

		copy(flatDb[i*rowLen:], v[:])
	}
	return &StaticDB{len(data), rowLen, flatDb}, err
}

type StaticDBParams struct {
	NRows  int
	RowLen int
}

func (p *StaticDBParams) NumRows() int {
	return p.NRows
}

func (db StaticDB) Params() *StaticDBParams {
	return &StaticDBParams{db.NumRows, db.RowLen}
}

func XorInto(a []byte, b []byte) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	fastxor.Bytes(a, a, b)
}
