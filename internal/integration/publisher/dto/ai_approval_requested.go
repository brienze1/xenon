package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const AIApprovalRequestedEvent = "ai:approval-requested"

type AIApprovalRequested struct {
	Environment AIApprovalEnvironment `json:"environment"`
	Id          string                `json:"id"`
	Client      string                `json:"client"`
	Method      string                `json:"method"`
	Url         string                `json:"url"`
	CreatedAt   string                `json:"created_at"`
}

func NewAIApprovalRequested(approval entity.AIApproval) AIApprovalRequested {
	return AIApprovalRequested{
		Id:          approval.ID,
		Client:      approval.Client,
		Method:      approval.Method,
		Url:         approval.RawURL,
		CreatedAt:   approval.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Environment: NewAIApprovalEnvironment(approval.Environment),
	}
}
