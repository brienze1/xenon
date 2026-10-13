package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ResolveAIApproval interface {
	// Resolve applies resolve to the approval atomically; it returns nil for an unknown id and false when resolve refused.
	Resolve(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, bool)
}
