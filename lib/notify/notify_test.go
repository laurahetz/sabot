package notify

import (
	"log"
	"math/rand"
	"sabot/lib/util"
	"slices"
	"testing"
)

func TestBuildMatrix(t *testing.T) {

	r := rand.New(rand.NewSource(42))
	size := 100 // TODO make bigger
	nM := NewMatrix(size)

	listTargets := make([][]uint32, size)
	targetMap := make(map[uint32][]uint32)

	for i := 0; i < size; i++ {
		numReceivers := rand.Intn(size)
		targets := util.RandTargets(r, numReceivers, size-1, 0)

		for _, target := range targets {
			targetMap[uint32(target)] = append(targetMap[uint32(target)], uint32(i))
		}
		col := CreateVector(targets[:], uint32(size))
		// check create vector and read vector contain same elements
		readTargets := ReadVector(col)

		// check written column and read column are the same
		// only for checking equality
		slices.Sort(targets)
		slices.Sort(readTargets)
		if len(targets) != len(readTargets) {
			t.Fatal("failed to correctly read or write all entries from/to vector")
		}
		for i := range targets {
			if targets[i] != readTargets[i] {
				t.Fatal("vector input and output not equal")
			}
		}
		// save all targets for receivers for later checks
		listTargets[i] = targets[:]
		// Set sender vector in notify matrix
		nM.SetColumn(i, col)
	}

	for receiver := 0; receiver < size; receiver++ {
		// get row with index i from matrix
		row := nM.GetRow(uint32(receiver))

		// get senders from row
		retrieved_senders := ReadVector(row)
		if len(targetMap[uint32(receiver)]) != len(retrieved_senders) {
			// log.Println("targetMap:\t", targetMap[uint32(receiver)])
			// log.Println("retrieved_senders\t:", retrieved_senders)
			log.Fatal("number of senders does not match for receiver ", receiver)
		}

		// if targetMap for receiver is empty and no senders have been retrieved
		if len(targetMap[uint32(receiver)]) == 0 {
			break
		}

		for _, sender := range targetMap[uint32(receiver)] {
			found := false
			for _, retrieved_sender := range retrieved_senders {
				if uint32(retrieved_sender) == uint32(sender) {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("retrieved sender ", sender, "for receiver ", receiver, "but receiver is not in senders targets: ", targetMap[uint32(receiver)])
			}
		}
	}
}
