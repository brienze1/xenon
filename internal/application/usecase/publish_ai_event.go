package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishAIEvent interface {
	AccessStatusChanged(ctx context.Context, status entity.AIAccessStatus) error
	ApprovalRequested(ctx context.Context, approval entity.AIApproval) error
	ApprovalResolved(ctx context.Context, approval entity.AIApproval) error
	HistoryChanged(ctx context.Context, localId string) error
	WorkspacesChanged(ctx context.Context, activeLocalId string) error
}
