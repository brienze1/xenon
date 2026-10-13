package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveExample struct {
	LocalId      string      `json:"local_id" validate:"required,not_blank"`
	RequestId    string      `json:"request_id" validate:"required,not_blank"`
	BaseRevision string      `json:"base_revision"`
	Example      ExampleEdit `json:"example"`
	Force        bool        `json:"force"`
}

func (s *SaveExample) ToEntity() entity.SavedExample {
	return s.Example.ToEntity()
}

func (s *SaveExample) KeepsBody() bool {
	return s.Example.Response.Body == nil
}

func (s SaveExample) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "request_id": s.RequestId, "example_id": s.Example.Id, "base_revision": s.BaseRevision, "force": s.Force}
}
