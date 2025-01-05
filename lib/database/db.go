package database

import (
	"bytes"
	"errors"
	"log"
	"math"
	"math/bits"
	"runtime"
	"sabot/lib/merkle"
	"sabot/lib/util"

	"github.com/spaolacci/murmur3"
)

const (
	MaxIterations = 1024
)

type DBParams struct {
	NRows              uint32
	Auth               bool // true if APIR is used
	Seed               uint64
	SegmentLength      uint32
	SegmentLengthMask  uint32
	SegmentCount       uint32
	SegmentCountLength uint32
	KeyLength          uint32
	ValueLength        uint32
	RecordLength       uint32 // KeyLength + ValueLength + ProofLen (+ 4 Byte if KWPIR DB)
	Root               []byte //for merkle proof
	ProofLen           uint32 //for merkle proof
}

type IKVElement struct {
	Idx   uint32
	Key   []byte
	Value []byte
}

type KVElement struct {
	Key   []byte
	Value []byte
}

type Database struct {
	Pp DBParams
	Db *StaticDB
}

/* From BFF implementation  */
func calculateSegmentLength(arity uint32, size uint32) uint32 {
	// These parameters are very sensitive. Replacing 'floor' by 'round' can
	// substantially affect the construction time.

	// if number of keys is 0
	if size == 0 {
		return 4
	}
	if arity == 3 {
		return uint32(1) << int(math.Floor(math.Log(float64(size))/math.Log(3.33)+2.25))
	} else if arity == 4 {
		return uint32(1) << int(math.Floor(math.Log(float64(size))/math.Log(2.91)-0.5))
	} else {
		return 65536
	}
}

/* From BFF implementation  */
func calculateSizeFactor(arity uint32, size uint32) float64 {
	if arity == 3 {
		return math.Max(1.125, 0.875+0.25*math.Log(1000000)/math.Log(float64(size)))
	} else if arity == 4 {
		return math.Max(1.075, 0.77+0.305*math.Log(600000)/math.Log(float64(size)))
	} else {
		return 2.0
	}
}

// Bootstrapping: Get Public Parameters and an empty temporary data store
// Code adapted from BFF, uses peeling setup
func initSetup(size uint32, arity uint32) (*DBParams, *[]IKVElement) {
	pp := &DBParams{}
	pp.SegmentLength = calculateSegmentLength(arity, size)

	if pp.SegmentLength > 262144 {
		pp.SegmentLength = 262144
	}
	pp.SegmentLengthMask = pp.SegmentLength - 1
	sizeFactor := calculateSizeFactor(arity, size)
	capacity := uint32(0)
	if size > 1 {
		capacity = uint32(math.Round(float64(size) * sizeFactor))
	}
	initSegmentCount := (capacity+pp.SegmentLength-1)/pp.SegmentLength - (arity - 1)
	arrayLength := (initSegmentCount + arity - 1) * pp.SegmentLength
	pp.SegmentCount = (arrayLength + pp.SegmentLength - 1) / pp.SegmentLength
	if pp.SegmentCount <= arity-1 {
		pp.SegmentCount = 1
	} else {
		pp.SegmentCount = pp.SegmentCount - (arity - 1)
	}
	arrayLength = (pp.SegmentCount + arity - 1) * pp.SegmentLength
	pp.SegmentCountLength = pp.SegmentCount * pp.SegmentLength

	// initialize empty filter values used in setup
	tempStore := make([]IKVElement, arrayLength)
	return pp, &tempStore
}

/*
from BFF implementation,
returns random number, modifies the seed
*/
func splitmix64(seed *uint64) uint64 {
	*seed = *seed + 0x9E3779B97F4A7C15
	z := *seed
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// Adapted from BFF implementation to support byte array instead of uint64 inputs
func mixsplit(key []byte, seed uint64) uint64 {
	return murmur3.Sum64WithSeed(key, uint32(seed))
}

// Adapted from BFF implementation
// Only works for ARITY = 3 !!!
func (pp *DBParams) getHashFromHash(hash uint64) (uint32, uint32, uint32) {
	hi, _ := bits.Mul64(hash, uint64(pp.SegmentCountLength))
	h0 := uint32(hi)
	h1 := h0 + pp.SegmentLength
	h2 := h1 + pp.SegmentLength
	h1 ^= uint32(hash>>18) & pp.SegmentLengthMask
	h2 ^= uint32(hash) & pp.SegmentLengthMask
	return h0, h1, h2
}

// From BFF implementation
func mod3(x uint8) uint8 {
	if x > 2 {
		x -= 3
	}
	return x
}

/*
SetupBinaryFuse creates a new tempDB  and fills it with provided key-values.
For best results, the input should not have (too many) duplicated keys.
The function may return an error if the set is empty.

Bootstrapping: Setup data store using Binary Fuse Peeling technique and parameters
The value here is not only the value, but key||value
and in the authenticate case it is idx||key||value or some slightly different order
*/
func SetupBinaryFuse(kvelements []KVElement, arity uint32) (pp *DBParams, bff *[]IKVElement, err error) {
	size := uint32(len(kvelements))
	pp, bff = initSetup(size, arity)
	rngcounter := uint64(1)
	pp.Seed = splitmix64(&rngcounter)

	capacity := uint32(len(*bff))
	if size == 0 {
		return pp, bff, nil
	}
	pp.KeyLength = uint32(len(kvelements[0].Key))
	pp.ValueLength = uint32(len(kvelements[0].Value))

	alone := make([]uint32, capacity)
	// BFF: the lowest 2 bits are the h index (0, 1, or 2)
	// so we only have 6 bits for counting; but that's sufficient
	t2count := make([]uint8, capacity)
	reverseH := make([]uint8, size)

	t2hash := make([]uint64, capacity)
	reverseOrder := make([]uint64, size+1)
	// Bootstrapping: needed to map hash to its kvpair
	hashDataMap := make(map[uint64]IKVElement, size+1)
	reverseOrder[size] = 1

	// BFF: the array h0, h1, h2, h0, h1, h2
	var h012 [6]uint32
	// BFF: this could be used to compute the mod3
	// tabmod3 := [5]uint8{0,1,2,0,1}
	iterations := 0
	for {
		iterations += 1
		if iterations > MaxIterations {
			// BFF: The probability of this happening is lower than the
			// the cosmic-ray probability (i.e., a cosmic ray corrupts your system).
			return pp, nil, errors.New("too many iterations")
		}

		blockBits := 1
		for (1 << blockBits) < pp.SegmentCount {
			blockBits += 1
		}
		startPos := make([]uint, 1<<blockBits)
		for i := range startPos {
			// BFF: important: we do not want i * size to overflow!!!
			startPos[i] = uint((uint64(i) * uint64(size)) >> blockBits)
		}
		for _, kvpair := range kvelements {
			hash := mixsplit(kvpair.Key, pp.Seed)
			segment_index := hash >> (64 - blockBits)
			for reverseOrder[startPos[segment_index]] != 0 {
				segment_index++
				segment_index &= (1 << blockBits) - 1
			}
			reverseOrder[startPos[segment_index]] = hash
			// Bootstrapping: remember which hash belongs to what final DB values, i.e. K||V
			hashDataMap[hash] = IKVElement{Key: kvpair.Key, Value: kvpair.Value}
			startPos[segment_index] += 1
		}
		error := 0
		for i := uint32(0); i < size; i++ {
			hash := reverseOrder[i]
			index1, index2, index3 := pp.getHashFromHash(hash)
			t2count[index1] += 4
			// BFF: t2count[index1] ^= 0 // noop
			t2hash[index1] ^= hash
			t2count[index2] += 4
			t2count[index2] ^= 1
			t2hash[index2] ^= hash
			t2count[index3] += 4
			t2count[index3] ^= 2
			t2hash[index3] ^= hash
			// BFF: If we have duplicated hash values, then it is likely that
			// the next comparison is true
			// Bootstrapping: We can't handle hash collisions! So we need to redo it with different hash functions
			if t2hash[index1]&t2hash[index2]&t2hash[index3] == 0 {
				// BFF: next we do the actual test
				if ((t2hash[index1] == 0) && (t2count[index1] == 8)) || ((t2hash[index2] == 0) && (t2count[index2] == 8)) || ((t2hash[index3] == 0) && (t2count[index3] == 8)) {
					error = 1
					break
				}
			}
			if t2count[index1] < 4 || t2count[index2] < 4 || t2count[index3] < 4 {
				error = 1
				break
			}
		}
		// Reset filter setup and retry
		if error == 1 {
			for i := uint32(0); i < size; i++ {
				reverseOrder[i] = 0
			}
			hashDataMap = make(map[uint64]IKVElement, size+1)
			for i := uint32(0); i < capacity; i++ {
				t2count[i] = 0
				t2hash[i] = 0
			}
			pp.Seed = splitmix64(&rngcounter)
			continue
		}

		// End of key addition
		Qsize := 0
		// Add sets with one key to the queue.
		for i := uint32(0); i < capacity; i++ {
			alone[Qsize] = i
			if (t2count[i] >> 2) == 1 {
				Qsize++
			}
		}
		stacksize := uint32(0)
		for Qsize > 0 {
			Qsize--
			index := alone[Qsize]
			if (t2count[index] >> 2) == 1 {
				hash := t2hash[index]
				found := t2count[index] & 3
				reverseH[stacksize] = found
				reverseOrder[stacksize] = hash
				stacksize++

				index1, index2, index3 := pp.getHashFromHash(hash)

				h012[1] = index2
				h012[2] = index3
				h012[3] = index1
				h012[4] = h012[1]

				other_index1 := h012[found+1]
				alone[Qsize] = other_index1
				if (t2count[other_index1] >> 2) == 2 {
					Qsize++
				}
				t2count[other_index1] -= 4
				t2count[other_index1] ^= mod3(found + 1) // could use this instead: tabmod3[found+1]
				t2hash[other_index1] ^= hash

				other_index2 := h012[found+2]
				alone[Qsize] = other_index2
				if (t2count[other_index2] >> 2) == 2 {
					Qsize++
				}
				t2count[other_index2] -= 4
				t2count[other_index2] ^= mod3(found + 2) // could use this instead: tabmod3[found+2]
				t2hash[other_index2] ^= hash
			}
		}
		break // Successfull  filter creation
	}

	for i := int(size - 1); i >= 0; i-- {
		// BFF: the hash of the key we insert next
		hash := reverseOrder[i]
		data := hashDataMap[hash]
		index1, index2, index3 := pp.getHashFromHash(hash)
		found := reverseH[i]
		h012[0] = index1
		h012[1] = index2
		h012[2] = index3
		h012[3] = h012[0]
		h012[4] = h012[1]

		// write IKV to primary index
		data.Idx = h012[found]
		(*bff)[h012[found]] = data
	}
	return pp, bff, err
}

// Returns all posible indices for a key in the tempDB
func (pp *DBParams) GetIndices(key []byte) []uint32 {
	// hash key using mixsplit() and compute indices based on that
	h0, h1, h2 := pp.getHashFromHash(mixsplit(key, pp.Seed))
	return []uint32{h0, h1, h2}
}

// Contains returns `true` if key is part of the set with a false positive probability of <0.4%.
func (db *Database) Contains(key []byte) bool {
	indices := db.Pp.GetIndices(key)
	for _, i := range indices {
		if len(db.Db.Row(int(i))) > 0 &&
			bytes.Equal(db.Db.Row(int(i))[:db.Pp.KeyLength], key) {
			return true
		}
	}
	return false
}

// Returns value for keyword, does not return proof in auth mode
func (db *Database) Get(key []byte) (bool, *IKVElement) {
	indices := db.Pp.GetIndices(key)
	for _, i := range indices {
		row := db.Db.Row(int(i))

		// check if key matches
		if len(row) > int(db.Pp.KeyLength) &&
			bytes.Equal(row[:db.Pp.KeyLength], key) {
			return true, &IKVElement{i, key, row[db.Pp.KeyLength:(db.Pp.KeyLength + db.Pp.ValueLength)]}
		}
	}
	return false, &IKVElement{}
}

func (pp *DBParams) RowToIKV(idx uint32, row []byte) IKVElement {
	return IKVElement{Idx: idx, Key: row[:pp.KeyLength], Value: row[pp.KeyLength:(pp.KeyLength + pp.ValueLength)]}
}

func (pp *DBParams) IKVToRow(ikv IKVElement) []byte {
	row := append(ikv.Key, ikv.Value...)
	if pp.Auth {
		row = append(row, util.Uint32ToByteSlice(ikv.Idx)...)
	}
	return row
}

func (pp *DBParams) VerifyRow(row []byte) ([]byte, error) {
	data := row[:len(row)-int(pp.ProofLen)]
	// check Merkle proof
	proof := merkle.DecodeProof(row[len(row)-int(pp.ProofLen):])
	verified, err := merkle.VerifyProof(data, proof, pp.Root)
	if err != nil {
		log.Fatalf("impossible to verify proof: %v", err)
	}
	if !verified {
		return nil, errors.New("reject proof")
	}
	return data, nil
}

// Updates tempDb to include merkle proofs and adds merkle params to PP
func (pp *DBParams) CreateMerkle(tempDB *[][]byte) {
	tree, err := merkle.New(*tempDB)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}
	// GC after tree generation
	runtime.GC()

	// Set Params
	pp.ProofLen = uint32(tree.EncodedProofLength())
	pp.Root = tree.Root()

	// tempDB entries are extended to include the merkle proofs
	for i, row := range *tempDB {
		p, err := tree.GenerateProof(row)
		if err != nil {
			log.Fatalf("error while generating proof for row %v: %v", i, err)
		}
		// appending the proof to the tempDB values
		(*tempDB)[i] = append((*tempDB)[i], merkle.EncodeProof(p)...)
	}
}
