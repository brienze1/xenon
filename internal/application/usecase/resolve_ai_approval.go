package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ResolveAIApproval interface {
	Resolve(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, error)
}
