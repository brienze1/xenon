package dto

type QuickSearch struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Query   string `json:"query"`
	Limit   int    `json:"limit" validate:"min=0"`
}
