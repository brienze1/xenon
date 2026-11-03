package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveAISessionContext interface {
	Save(ctx context.Context, sessionContext entity.AISessionContext)
}
