package pir

import (
	"fmt"
	"math"
	"math/rand"
	"sabot/lib/database"

	"github.com/dkales/dpf-go/dpf"
)

const (
	Left  int = 0
	Right int = 1
)

type DpfClient struct {
	*database.StaticDBParams
	randSource *rand.Rand
}

type DPFQueryResp struct {
	Answer []byte
}

type ReconstructFunc func(resp []interface{}) ([]byte, error)

func InitPIRClient(dbParams *database.StaticDBParams, source *rand.Rand) *DpfClient {
	return &DpfClient{dbParams, source}
}

func (c *DpfClient) Query(idx int) ([]*dpf.DPFkey, ReconstructFunc) {
	numBits := uint64(math.Ceil(math.Log2(float64(c.NRows))))
	qL, qR := dpf.Gen(uint64(idx), numBits)

	return []*dpf.DPFkey{&qL, &qR}, func(resps []interface{}) ([]byte, error) {
		queryResps := make([]*DPFQueryResp, len(resps))
		var ok bool
		for i, r := range resps {
			if queryResps[i], ok = r.(*DPFQueryResp); !ok {
				return nil, fmt.Errorf("Invalid response type: %T, expected *DPFQueryResp", r)
			}
		}

		return c.reconstruct(queryResps)
	}
}

func (c *DpfClient) DummyQuery() []*dpf.DPFkey {
	q, _ := c.Query(0)
	return q
}

func (c *DpfClient) reconstruct(resp []*DPFQueryResp) ([]byte, error) {
	out := make([]byte, len(resp[Left].Answer))
	database.XorInto(out, resp[Left].Answer)
	database.XorInto(out, resp[Right].Answer)
	return out, nil
}

func (c *DpfClient) Reconstruct(resp [][]byte) ([]byte, error) {
	out := make([]byte, len(resp[Left]))
	database.XorInto(out, resp[Left])
	database.XorInto(out, resp[Right])
	return out, nil
}

func matVecProduct(db *database.StaticDB, bitVector []byte) []byte {
	out := make([]byte, db.RowLen)
	if db.RowLen == 32 {
		XorHashesByBitVector(db.FlatDb, bitVector, out)
	} else {
		var j uint
		for j = 0; j < uint(db.NumRows); j++ {
			if ((1 << (j % 8)) & bitVector[j/8]) != 0 {
				database.XorInto(out, db.FlatDb[j*uint(db.RowLen):(j+1)*uint(db.RowLen)])
			}
		}
	}
	return out
}

func Process(db *database.StaticDB, key *dpf.DPFkey) (*DPFQueryResp, error) {
	bitVec := dpf.EvalFull(*key, uint64(math.Ceil(math.Log2(float64(db.NumRows)))))
	return &DPFQueryResp{matVecProduct(db, bitVec)}, nil
}

func Process_old(db *database.StaticDB, key *dpf.DPFkey) (interface{}, error) {
	bitVec := dpf.EvalFull(*key, uint64(math.Ceil(math.Log2(float64(db.NumRows)))))
	return &DPFQueryResp{matVecProduct(db, bitVec)}, nil
}
