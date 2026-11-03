package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAISessionContext interface {
	Find(ctx context.Context, sessionId string) (entity.AISessionContext, bool)
}
