package dto

type SaveRequestAs struct {
	LocalId  string  `json:"local_id" validate:"required,not_blank"`
	ParentId string  `json:"parent_id" validate:"required,not_blank"`
	Name     string  `json:"name" validate:"required,not_blank,max=200"`
	Request  Request `json:"request"`
}

func (s SaveRequestAs) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "parent_id": s.ParentId, "name": s.Name, "request": s.Request.LogValue()}
}
