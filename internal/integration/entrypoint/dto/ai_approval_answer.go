package dto

type AIApprovalAnswer struct {
	Id       string `json:"id" validate:"required,not_blank"`
	Approved bool   `json:"approved"`
}
