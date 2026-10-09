package dto

type PickFolder struct {
	Title string `json:"title" validate:"required,not_blank"`
}
