package dto

type CreateEnvironment struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Name    string `json:"name" validate:"required,not_blank"`
}
