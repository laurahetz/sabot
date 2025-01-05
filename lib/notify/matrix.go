package notify

import (
	"sabot/lib/database"
)

type NotifyMatrix struct {
	*database.StaticDB
}

func (nM *NotifyMatrix) SetColumn(cIdx int, col []byte) {
	// need to transform col (where each bit indicates a relation) to a matrix column where each
	// byte indicates a relation. Bytes are the smalles unit in go, so getting the row should be easiest this way
	rIndices := ReadVector(col)

	for _, rIdx := range rIndices {
		nM.FlatDb[int(rIdx)*nM.RowLen+cIdx] = byte(1)
	}
}

// returns row in compressed byte array, where one bit represents a relation
func (nM *NotifyMatrix) GetRow(rIdx uint32) []byte {
	if rIdx < uint32(nM.NumRows) {
		matrixRow := nM.Row(int(rIdx))
		return VecFromMatrixSlice(matrixRow)
	}
	return nil
}

func NewMatrix(size int) *NotifyMatrix {
	nM := NotifyMatrix{}
	nM.StaticDB = &database.StaticDB{}
	nM.RowLen = size
	nM.NumRows = size
	nM.FlatDb = make([]byte, size*size)
	// ensures Matrix is properly initialized and actually using memory
	for i := range nM.FlatDb {
		nM.FlatDb[i] = 0
	}
	return &nM
}
