package dto

type CreateAlias struct {
	LocalId      string `json:"local_id" validate:"required,not_blank"`
	RequestId    string `json:"request_id" validate:"required,not_blank"`
	BaseRevision string `json:"base_revision"`
	Name         string `json:"name"`
}
