package dto

type SearchTree struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Query   string `json:"query"`
}
