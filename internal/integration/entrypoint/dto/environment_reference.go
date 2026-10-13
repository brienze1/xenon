package dto

type EnvironmentReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Id      string `json:"id" validate:"required,not_blank"`
}
