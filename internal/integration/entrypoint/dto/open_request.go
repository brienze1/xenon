package dto

type OpenRequest struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Id      string `json:"id" validate:"required,not_blank"`
}
