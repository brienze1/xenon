package dto

type AliasReference struct {
	LocalId      string `json:"local_id" validate:"required,not_blank"`
	RequestId    string `json:"request_id" validate:"required,not_blank"`
	AliasId      string `json:"alias_id" validate:"required,not_blank"`
	BaseRevision string `json:"base_revision"`
}
