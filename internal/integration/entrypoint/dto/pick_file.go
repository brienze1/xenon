package dto

type PickFile struct {
	Title string `json:"title" validate:"required,not_blank"`
}
