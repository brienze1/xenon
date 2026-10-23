package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApproval struct {
	Environment AIApprovalEnvironment `json:"environment"`
	Id          string                `json:"id"`
	Client      string                `json:"client"`
	Method      string                `json:"method"`
	Url         string                `json:"url"`
	Summary     string                `json:"summary"`
	CreatedAt   string                `json:"created_at"`
}

func NewAIApproval(approval entity.AIApproval) AIApproval {
	return AIApproval{
		Id:          approval.ID,
		Client:      approval.Client,
		Method:      approval.Method,
		Url:         approval.RawURL,
		Summary:     approval.Summary,
		CreatedAt:   approval.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Environment: NewAIApprovalEnvironment(approval.Environment),
	}
}

func NewAIApprovals(approvals []entity.AIApproval) []AIApproval {
	rows := make([]AIApproval, 0, len(approvals))
	for _, approval := range approvals {
		rows = append(rows, NewAIApproval(approval))
	}
	return rows
}

func (a AIApproval) LogValue() any {
	return map[string]any{"id": a.Id, "client": a.Client, "method": a.Method, "environment_id": a.Environment.Id, "created_at": a.CreatedAt}
}
