package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type AIAccessController interface {
	Status(ctx context.Context) (*dto.AIAccessStatus, error)
	Start(ctx context.Context) (*dto.AIAccessStatus, error)
	Stop(ctx context.Context) (*dto.AIAccessStatus, error)
	Guard(ctx context.Context, request dto.AIGuard) (*dto.AIGuarded, error)
	Notify(ctx context.Context, request dto.AIChanges) error
	Record(ctx context.Context, request dto.AIActivityRecord) error
	RequestApproval(ctx context.Context, request dto.AIApprovalRequest) (*dto.AIApprovalOutcome, error)
	AnswerApproval(ctx context.Context, request dto.AIApprovalAnswer) error
	PendingApprovals(ctx context.Context) ([]dto.AIApproval, error)
	Shutdown(ctx context.Context) error
	ClientsChanged(ctx context.Context) error
}
