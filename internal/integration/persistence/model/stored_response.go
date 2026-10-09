package model

import "github.com/brienze1/xenon/internal/domain/entity"

type StoredResponse struct {
	Response entity.Response
	BodyPath string
}

func NewStoredResponse(response entity.Response) StoredResponse {
	return StoredResponse{Response: response}
}

func NewSpilledResponse(response entity.Response, bodyPath string) StoredResponse {
	return StoredResponse{Response: response.WithoutBody(), BodyPath: bodyPath}
}

func (s *StoredResponse) ToEntity() entity.Response {
	return s.Response
}

func (s *StoredResponse) HasBodyFile() bool {
	return s.BodyPath != ""
}
