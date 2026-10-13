package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAIApprovals interface {
	FindPending(ctx context.Context) []entity.AIApproval
	FindSessionPending(ctx context.Context, sessionId string) []entity.AIApproval
}
