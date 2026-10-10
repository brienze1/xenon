package dto

type ListEnvironments struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
