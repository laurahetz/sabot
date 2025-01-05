package database

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sabot/lib/util"
	pb "sabot/proto/bootstrapping"

	"google.golang.org/protobuf/proto"
)

func GetTestData(numKeys uint32, keyLength uint, valLength uint, seed int64) (elements []KVElement) {
	elements = make([]KVElement, numKeys)

	r := rand.New(rand.NewSource(seed))
	for i := range elements {
		nKey := make([]byte, keyLength)
		nVal := make([]byte, valLength)

		_, err := r.Read(nKey)
		if err != nil {
			log.Fatalln("Error getting random bytes:", err)
		}
		_, err = r.Read(nVal)
		if err != nil {
			log.Fatalln("Error getting random bytes:", err)
		}
		elements[i].Key = nKey
		elements[i].Value = nVal
	}
	return
}

func RandTargetsExcept(r *rand.Rand, num int, max int, min int, exclude uint32) []uint32 {
	targets := make([]uint32, num)
	uM := make(map[uint32]bool)
	t := exclude
	for i := range targets {
		for {
			for {
				t = uint32(r.Intn(max+1-min) + min)
				if t != exclude {
					break
				}
			}
			if !uM[t] {
				uM[t] = true
				targets[i] = t
				break
			}
		}
	}
	return targets
}

func RandTargets(r *rand.Rand, num int, max int, min int) []uint32 {
	targets := make([]uint32, num)
	uM := make(map[uint32]bool)
	for i := range targets {
		for {
			t := uint32(r.Intn(max+1-min) + min)
			if !uM[t] {
				uM[t] = true
				targets[i] = t
				break
			}
		}
	}
	return targets
}

func GetTestKWs(numKeys uint32, keyLength uint, seed int64) (kws [][]byte) {
	kws = make([][]byte, numKeys)

	r := rand.New(rand.NewSource(seed))
	for i := range kws {
		nKey := make([]byte, keyLength)
		_, err := r.Read(nKey)
		if err != nil {
			log.Fatalln("Error getting random bytes:", err)
		}
		kws[i] = nKey
	}
	return
}

func EqualPublicParams(pp1 *DBParams, pp2 *DBParams) bool {
	if pp1.NRows != pp2.NRows {
		log.Println("NRows not equal")
		return false
	}
	if pp1.Auth != pp2.Auth {
		log.Println("Auth not equal")
		return false
	}
	if pp1.Seed != pp2.Seed {
		log.Println("Seed not equal")
		return false
	}
	if pp1.SegmentLength != pp2.SegmentLength {
		log.Println("SegmentLength not equal")
		return false
	}
	if pp1.SegmentLengthMask != pp2.SegmentLengthMask {
		log.Println("SegmentLengthMask not equal")
		return false
	}
	if pp1.SegmentCount != pp2.SegmentCount {
		log.Println("SegmentCount not equal")
		return false
	}
	if pp1.SegmentCountLength != pp2.SegmentCountLength {
		log.Println("SegmentCountLength not equal")
		return false
	}
	if pp1.KeyLength != pp2.KeyLength {
		log.Println("KeyLen not equal")
		return false
	}
	if pp1.ValueLength != pp2.ValueLength {
		log.Println("ValLen not equal")
		return false
	}
	if pp1.ProofLen != pp2.ProofLen {
		log.Println("ProofLen not equal")
		return false
	}
	if pp1.RecordLength != pp2.RecordLength {
		log.Println("RecordLength not equal")
		return false
	}
	if !bytes.Equal(pp1.Root, pp2.Root) {
		log.Println("Roots not equal")
		return false
	}
	return true
}

/* Stores PublicParams in one file and StaticDB in another */
func ContactDBToFile(path string, db *Database) {

	// encode PublicParams
	paramsProto := pb.Params{
		Nrows:       db.Pp.NRows,
		Auth:        db.Pp.Auth,
		Seed:        db.Pp.Seed,
		SegLen:      db.Pp.SegmentLength,
		SegLenMask:  db.Pp.SegmentLengthMask,
		SegCount:    db.Pp.SegmentCount,
		SegCountLen: db.Pp.SegmentCountLength,
		KeyLen:      db.Pp.KeyLength,
		ValLen:      db.Pp.ValueLength,
		ProofLen:    db.Pp.ProofLen,
		Root:        db.Pp.Root,
		RecLength:   db.Pp.RecordLength,
	}
	paramsEnc, err := proto.Marshal(&paramsProto)
	if err != nil {
		log.Fatal("error encoding public params:", err)
	}
	bytesToWrite := append(util.Uint32ToByteSlice(uint32(len(paramsEnc))), paramsEnc...)

	// open output file
	fo, err := os.Create(path)
	if err != nil {
		log.Fatal("error reading db file:", err)
	}
	defer fo.Close()

	w := bufio.NewWriter(fo)
	nBuf := 1024 * 4

	nParamChunks := int(math.Ceil(float64(len(bytesToWrite)) / float64(nBuf)))
	var chunkToWrite []byte
	for i := 0; i < nParamChunks; i++ {
		if i == nParamChunks-1 {
			padding := make([]byte, nBuf-len(bytesToWrite[i*nBuf:]))
			chunkToWrite = append(bytesToWrite[i*nBuf:], padding...)
		} else {
			chunkToWrite = bytesToWrite[i*nBuf : (i+1)*nBuf]
		}

		_, err = w.Write(chunkToWrite) // write a chunk
		if err != nil {
			log.Fatalln("error writing params bytes: ", err)
		}
		if err = w.Flush(); err != nil {
			log.Fatalln("error flushing writer:", err)
		}
	}

	// Write DB to following chunks
	start := 0
	end := 0
	for {
		if len(db.Db.FlatDb[start:])/nBuf > 1 {
			end += nBuf
		} else {
			end += len(db.Db.FlatDb[start:])
		}
		_, err = w.Write(db.Db.FlatDb[start:end])
		// write a chunk
		if err != nil {
			log.Fatalln("error writing bytes: ", err)
		}
		start = end
		if start > len(db.Db.FlatDb)-1 {
			break
		}
	}
	if err = w.Flush(); err != nil {
		log.Fatalln("error flushing writer:", err)
	}

}

func ContactDBFromFile(path string) (out_db *Database) {

	out_db = &Database{}
	nBuf := 4 * 1024
	fd, err := os.Open(path)
	if err != nil {
		log.Fatalln("Could not open path ", path)
	}
	defer fd.Close()

	b := make([]byte, nBuf)
	r := bufio.NewReader(fd)

	_, err = r.Read(b)
	if err != nil {
		log.Fatal("error reading db file:", err)
	}
	// Read size of Public Params
	paramsLen := int(binary.BigEndian.Uint32(b[:4]))
	ppBuf := make([]byte, paramsLen)

	// get number of chunks in which public params are stored
	nParamChunks := int(math.Ceil(float64(paramsLen+4) / float64(nBuf)))
	startRead := 4 // for i = 0 case
	startWrite := 0
	// read chunks
	for i := 0; i < nParamChunks; i++ {
		if i != 0 {
			_, err := r.Read(b)
			if err != nil {
				log.Fatal("error reading db file:", err)
			}
			startRead = 0
		}
		copy(ppBuf[startWrite:], b[startRead:])
		startWrite += len(b[startRead:])
		runtime.GC()
	}

	protoParams := &pb.Params{}
	if err := proto.Unmarshal(ppBuf, protoParams); err != nil {
		log.Fatalln("Failed to parse params:", err)
	}

	out_db.Pp = DBParams{
		NRows:              protoParams.Nrows,
		Auth:               protoParams.Auth,
		Seed:               protoParams.Seed,
		SegmentLength:      protoParams.SegLen,
		SegmentLengthMask:  protoParams.SegLenMask,
		SegmentCount:       protoParams.SegCount,
		SegmentCountLength: protoParams.SegCountLen,
		KeyLength:          protoParams.KeyLen,
		ValueLength:        protoParams.ValLen,
		ProofLen:           protoParams.ProofLen,
		Root:               protoParams.Root,
		RecordLength:       protoParams.RecLength,
	}
	out_db.Db = &StaticDB{}
	out_db.Db.NumRows = int(protoParams.Nrows)
	//out_db.Db.RowLen = int(protoParams.RecLength)
	out_db.Db.RowLen = int(out_db.Pp.RecordLength + out_db.Pp.ProofLen)
	//out_db.Db.RowLen = int(out_db.Pp.KeyLength + out_db.Pp.ValueLength + out_db.Pp.ProofLen)
	out_db.Db.FlatDb = make([]byte, out_db.Pp.NRows*uint32(out_db.Pp.RecordLength+out_db.Pp.ProofLen))

	writeStart := 0
	totalDBRead := 0
	nDBChunks := int(math.Ceil(float64(len(out_db.Db.FlatDb)) / float64(nBuf)))
	for i := 0; i < nDBChunks; i++ {
		// log.Println("chunk ", i)
		n, err := r.Read(b)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("error reading db file:", err)
		}
		temp := copy(out_db.Db.FlatDb[writeStart:], b[:n])
		writeStart = writeStart + temp
		totalDBRead += n
		//runtime.GC()
		if totalDBRead == len(out_db.Db.FlatDb) {
			break
		}
	}
	return out_db
}
