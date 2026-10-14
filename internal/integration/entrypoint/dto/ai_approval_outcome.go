package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApprovalOutcome struct {
	Id      string `json:"id"`
	Outcome string `json:"outcome"`
}

func NewAIApprovalOutcome(approval entity.AIApproval) AIApprovalOutcome {
	return AIApprovalOutcome{Id: approval.ID, Outcome: approval.Outcome.String()}
}

func (a AIApprovalOutcome) LogValue() any {
	return map[string]any{"id": a.Id, "outcome": a.Outcome}
}
