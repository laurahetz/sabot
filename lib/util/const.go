package util

import "time"

const (
	CERT_C_PATH_PRE = "cert/client"
	CERT_S_PATH_PRE = "cert/server"
	CERT_CA_PATH    = "cert/ca-cert.pem"
	DPF_SEED        = 41
	MAX_MSG_SIZE    = 1024 * 1024 * 64
	TIMEOUT         = 100 * time.Minute
	INPUT_SEED      = 42
	ARITY           = 3  //BFF setup param, = num hash funcs
	KEY_LENGTH      = 32 //size in byte of client identifier
	VAL_LENGTH      = 32 // size in byte of contact info
)
