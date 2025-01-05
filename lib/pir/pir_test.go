package pir

import (
	"reflect"
	"testing"
)

func TestDPF(t *testing.T) {
	db := MakeDB(512, 32)
	if db == nil {
		t.Fatal("error making test db")
	}
	i := 128
	client := InitPIRClient(db.Params(), RandSource())

	queryReq, reconstructFunc := client.Query(i)
	if reconstructFunc == nil {
		t.Fatalf("Failed to query: %d", i)
	}
	responses := make([]interface{}, 2)
	var err error
	responses[0], err = Process(db, queryReq[Left])
	if err != nil {
		t.Fatalf("server 1 failed to answer: %d", err)
	}

	responses[1], err = Process(db, queryReq[Right])
	if err != nil {
		t.Fatalf("server 1 failed to answer: %d", err)
	}
	res, err := reconstructFunc(responses)
	if err != nil {
		t.Fatalf("failed to reconstruct answer")
	}

	if !reflect.DeepEqual(res, db.Row(128)) {
		t.Fatal("retrieved element does not match")
	}
}
