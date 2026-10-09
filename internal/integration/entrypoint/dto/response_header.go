package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ResponseHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewResponseHeader(header entity.KeyValue) ResponseHeader {
	return ResponseHeader{Key: header.Key, Value: header.Value}
}
