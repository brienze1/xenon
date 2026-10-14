package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type AIApprovalPersistence interface {
	usecase.QueueAIApproval
	usecase.FindAIApprovals
	usecase.ResolveAIApproval
	usecase.SettleAIApproval
}
