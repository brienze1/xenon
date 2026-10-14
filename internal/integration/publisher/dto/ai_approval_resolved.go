package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const AIApprovalResolvedEvent = "ai:approval-resolved"

type AIApprovalResolved struct {
	Id      string `json:"id"`
	Outcome string `json:"outcome"`
}

func NewAIApprovalResolved(approval entity.AIApproval) AIApprovalResolved {
	return AIApprovalResolved{Id: approval.ID, Outcome: approval.Outcome.String()}
}
