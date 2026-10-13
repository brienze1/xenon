package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiExample struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Value     string `json:"value"`
	Binary    bool   `json:"binary"`
}

func NewApiExample(example entity.ApiExample) ApiExample {
	return ApiExample{Id: example.ID, Name: example.Name, MediaType: example.MediaType, Value: example.Value, Binary: example.Binary}
}
