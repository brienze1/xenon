package dto

type ClearHistory struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
