package dto

type RunRowReference struct {
	Index   *int   `json:"index" validate:"required,min=0"`
	LocalId string `json:"local_id" validate:"required,not_blank"`
	RunId   string `json:"run_id" validate:"required,not_blank"`
}
