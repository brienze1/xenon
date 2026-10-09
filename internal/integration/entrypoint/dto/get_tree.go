package dto

type GetTree struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
