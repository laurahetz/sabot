package notify

import (
	"math/rand"

	"github.com/lukechampine/fastxor"
)

func GenShares(input []byte, numShares int, seed int64) [][]byte {
	shares := make([][]byte, numShares)
	r := rand.New(rand.NewSource(seed))

	// numShares - 1 random shares
	for i := 0; i < numShares-1; i++ {
		s := make([]byte, len(input))
		r.Read(s)
		shares[i] = s
	}
	// last share is XOR of all previous shares and input
	finalShare := input
	for i := 0; i < numShares-1; i++ {
		fastxor.Bytes(finalShare, finalShare, shares[i])
	}
	shares[numShares-1] = finalShare

	return shares

}

func CombineShares(shares [][]byte) []byte {
	for _, share := range shares[1:] {
		fastxor.Bytes(shares[0], shares[0], share)
	}
	return shares[0]
}
