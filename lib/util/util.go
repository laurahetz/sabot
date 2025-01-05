package util

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"log"
	"math/rand"
	"os"
	"strconv"

	"google.golang.org/grpc/credentials"
)

func StringToUint32Slice(in []string) []uint32 {
	ret := make([]uint32, len(in))
	for i, str := range in {
		oneint, _ := strconv.Atoi(str)
		ret[i] = uint32(oneint)
	}
	return ret
}

func Uint32ToByteSlice(in uint32) []byte {
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, in)
	return out
}

func ByteSliceToUint32(in []byte) uint32 {
	if len(in)/4 != 1 {
		log.Fatal("byte array has wrong size, expected len=4, got len=", len(in))
	}
	return binary.BigEndian.Uint32(in)
}

// if isClient: set up credentials for the client, else for server
func LoadTLSCred(filePre string, caPath string, isClient bool) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(filePre+"-cert.pem", filePre+"-key.pem")
	if err != nil {
		return nil, err
	}

	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, err
	}

	var config *tls.Config
	if isClient {
		config = &tls.Config{
			ServerName:   "localhost",
			Certificates: []tls.Certificate{cert},
			RootCAs:      ca,
		}
	} else {
		config = &tls.Config{
			ClientAuth:   tls.RequireAnyClientCert,
			Certificates: []tls.Certificate{cert},
			ClientCAs:    ca,
		}
	}

	return credentials.NewTLS(config), nil
}

func ByteSliceToUint64(in []byte) uint64 {
	if len(in)/8 != 1 {
		log.Fatal("byte array too large")
	}
	return binary.BigEndian.Uint64(in)
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
