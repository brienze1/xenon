package dto

type AutoHeadersRequest struct {
	LocalId string  `json:"local_id" validate:"required"`
	Request Request `json:"request"`
}

func (a AutoHeadersRequest) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "request": a.Request.LogValue()}
}
