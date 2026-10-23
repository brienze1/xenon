package dto

type RunReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	RunId   string `json:"run_id" validate:"required,not_blank"`
}
