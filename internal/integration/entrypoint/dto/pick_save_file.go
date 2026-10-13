package dto

type PickSaveFile struct {
	Title         string `json:"title" validate:"required,not_blank"`
	SuggestedName string `json:"suggested_name"`
}
