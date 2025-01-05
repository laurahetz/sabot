package pir

/*
#cgo amd64 CXXFLAGS: -msse2 -msse -march=native -maes -Ofast -std=c++11
#cgo arm64 CXXFLAGS: -march=armv8-a+fp+simd+crypto+crc -Ofast -std=c++11
#cgo LDFLAGS: -static-libstdc++
#include "xor.h"
*/
import "C"
import (
	"unsafe"
)

func XorBlocks(db []byte, offsets []int, out []byte) {
	C.xor_rows((*C.uchar)(&db[0]), C.uint(len(db)), (*C.ulonglong)(unsafe.Pointer(&offsets[0])), C.uint(len(offsets)), C.uint(len(out)), (*C.uchar)(&out[0]))
}

func XorHashesByBitVector(db []byte, indexing []byte, out []byte) {
	C.xor_hashes_by_bit_vector((*C.uchar)(&db[0]), C.uint(len(db)),
		(*C.uchar)(&indexing[0]), (*C.uchar)(&out[0]))
}
