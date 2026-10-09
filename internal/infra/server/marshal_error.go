package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/brienze1/xenon/pkg/errs"
)

func MarshalError(err error) []byte {
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		custom = errs.InternalServerError(err, "UNK-00500", "Unknown error")
	}
	payload, _ := json.Marshal(custom.ToDto(transactionId()))
	return payload
}

func transactionId() string {
	buffer := make([]byte, 16)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
