package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ServeAIAccess interface {
	Apply(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus
	Status(ctx context.Context) entity.AIAccessStatus
	Stop(ctx context.Context) entity.AIAccessStatus
}
