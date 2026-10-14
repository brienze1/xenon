package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAIAccessStatus interface {
	Status(ctx context.Context) entity.AIAccessStatus
}
