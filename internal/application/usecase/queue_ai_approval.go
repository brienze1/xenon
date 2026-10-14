package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type QueueAIApproval interface {
	// Queue keeps the approval pending for its session; the returned channel receives it once, when it is resolved.
	Queue(ctx context.Context, approval entity.AIApproval, sessionId string) <-chan entity.AIApproval
}
