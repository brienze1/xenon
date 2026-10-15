package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveAlias struct {
	LocalId      string       `json:"local_id" validate:"required,not_blank"`
	RequestId    string       `json:"request_id" validate:"required,not_blank"`
	BaseRevision string       `json:"base_revision"`
	Alias        RequestAlias `json:"alias"`
	Force        bool         `json:"force"`
}

func NewSaveAlias(localId, requestId, baseRevision string, alias RequestAlias, force bool) SaveAlias {
	return SaveAlias{LocalId: localId, RequestId: requestId, BaseRevision: baseRevision, Alias: alias, Force: force}
}

func (s *SaveAlias) ToEntity() entity.RequestAlias {
	return s.Alias.ToEntity()
}

func (s SaveAlias) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "request_id": s.RequestId, "alias": s.Alias.LogValue(), "base_revision": s.BaseRevision, "force": s.Force}
}
