package dto

type SaveRequest struct {
	LocalId      string  `json:"local_id" validate:"required,not_blank"`
	BaseRevision string  `json:"base_revision"`
	Request      Request `json:"request" validate:"has_id"`
	Force        bool    `json:"force"`
}

func (s SaveRequest) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "base_revision": s.BaseRevision, "force": s.Force, "request": s.Request.LogValue()}
}
