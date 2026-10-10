package dto

type SuggestVariables struct {
	LocalId   string `json:"local_id" validate:"required,not_blank"`
	RequestId string `json:"request_id"`
	Prefix    string `json:"prefix"`
}
