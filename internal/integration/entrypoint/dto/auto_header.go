package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AutoHeader struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

func NewAutoHeader(header entity.AutoHeader) AutoHeader {
	return AutoHeader{Key: header.Key, Value: header.Value, Disabled: header.Disabled}
}
