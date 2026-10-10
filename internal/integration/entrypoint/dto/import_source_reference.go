package dto

type ImportSourceReference struct {
	SourceId string `json:"source_id" validate:"required,not_blank"`
}
