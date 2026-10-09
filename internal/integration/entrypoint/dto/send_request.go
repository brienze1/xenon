package dto

type SendRequest struct {
	LocalId string  `json:"local_id" validate:"required"`
	TabId   string  `json:"tab_id" validate:"required"`
	Request Request `json:"request"`
}

func (s SendRequest) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "tab_id": s.TabId, "request": s.Request.LogValue()}
}
