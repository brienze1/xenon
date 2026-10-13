package dto

type ConsoleEntryReference struct {
	Id string `json:"id" validate:"required,not_blank"`
}
