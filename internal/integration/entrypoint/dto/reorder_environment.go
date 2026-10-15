package dto

type ReorderEnvironment struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Id      string `json:"id" validate:"required,not_blank"`
	Index   int    `json:"index" validate:"min=0"`
}
