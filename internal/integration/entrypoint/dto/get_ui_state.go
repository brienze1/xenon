package dto

type GetUIState struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
