package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAIAccessService interface {
	Status(ctx context.Context) entity.AIAccessStatus
	Start(ctx context.Context) (*entity.AIAccessStatus, error)
	Stop(ctx context.Context) entity.AIAccessStatus
}
