package dto

type GitAccountReference struct {
	ID string `json:"id" validate:"required,not_blank"`
}
