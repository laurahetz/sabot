package notify

import (
	"log"
	"math"
	"sabot/lib/database"
)

/*
Creates a notification vector
size: number of bits
targets: list of indices which should be 1
bootstrapping.Client has own impl of this function based on stored contacts
*/
func CreateVectorIKV(targets *[]database.IKVElement, size uint32) []byte {
	numBytes := uint32(math.Ceil(float64(size) / 8))
	col := make([]byte, numBytes)
	for _, target := range *targets {
		if target.Idx >= uint32(size) {
			log.Fatal("target ", target, " could not be added")
		} else {
			offset := int(math.Floor(float64(target.Idx) / 8))
			mask := byte(128) >> (target.Idx % 8)
			col[offset] = col[offset] ^ mask
		}
	}
	return col
}

func VecFromMatrixSlice(mSlice []byte) []byte {
	var targets []uint32
	for i, entry := range mSlice {
		if entry == byte(1) {
			targets = append(targets, uint32(i))
		}
	}
	return CreateVector(targets, uint32(len(mSlice)))
}

// Same as CreateVectorIKV, but has list of uints as input instead of IKV
func CreateVector(targets []uint32, size uint32) []byte {
	numBytes := uint32(math.Ceil(float64(size) / 8))
	col := make([]byte, numBytes)
	for _, target := range targets {
		if target >= uint32(size) {
			log.Fatal("target ", target, " could not be added")
		} else {
			offset := int(math.Floor(float64(target) / 8))
			mask := byte(128) >> (target % 8)
			col[offset] = col[offset] ^ mask
		}
	}
	return col
}

// return a list of (bit) indices where vector is 1
func ReadVector(row []byte) []uint32 {
	var senders []uint32
	for index, val := range row {
		offset := 7 + 8*index
		for i := 7; i >= 0; i-- {
			if (val>>i)&1 == 1 {
				senders = append(senders, uint32(offset-i))
			}
		}
	}
	return senders
}
