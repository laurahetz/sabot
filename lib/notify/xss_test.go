package notify

import (
	"math/rand"
	"reflect"
	"sabot/lib/util"
	"testing"
)

func TestGenAndCombineShares(t *testing.T) {

	r := rand.New(rand.NewSource(42))

	numShares := 2
	var seed int64 = 42
	var size uint32 = 10
	input := util.RandTargets(r, 2, int(size-1), 0)

	colVec := CreateVector(input, size)
	readColVec := ReadVector(colVec)

	if len(input) != len(readColVec) {
		t.Fatal("failed to correctly read or write all entries from/to vector")
	}
	foundCheck := make([]bool, size)
	for _, inEl := range input {
		for _, outEl := range readColVec {
			if inEl == outEl {
				foundCheck[inEl] = true
				break
			}
		}
		if !foundCheck[inEl] {
			t.Fatal("target was read that was not written!")
		}
	}
	for _, outEl := range readColVec {
		if !foundCheck[outEl] {
			t.Fatal("target was written but was not read!")
		}
	}

	// Gen shares
	shares := GenShares(colVec, numShares, seed)

	// Combine shares
	outVec := CombineShares(shares)

	output := ReadVector(outVec)

	if !reflect.DeepEqual(input, output) {
		t.Fatal("input and output are not equal")
	}
}
