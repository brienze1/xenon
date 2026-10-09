package dto

type ApplyRequest struct {
	Request Request `json:"request"`
}

func (a ApplyRequest) LogValue() any {
	return a.Request.LogValue()
}
