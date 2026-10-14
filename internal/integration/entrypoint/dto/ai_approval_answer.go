package dto

type AIApprovalAnswer struct {
	Id       string `json:"id" validate:"required,not_blank"`
	Approved bool   `json:"approved"`
}

func (a *AIApprovalAnswer) ApprovalId() string {
	return a.Id
}

func (a *AIApprovalAnswer) IsApproved() bool {
	return a.Approved
}
