package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAIAccessStatusService interface {
	Find(ctx context.Context) entity.AIAccessStatus
}
