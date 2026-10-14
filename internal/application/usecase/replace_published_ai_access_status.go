package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReplacePublishedAIAccessStatus interface {
	ReplacePublished(ctx context.Context, status entity.AIAccessStatus) entity.AIAccessStatus
}
