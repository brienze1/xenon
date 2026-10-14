package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAIAccessService interface {
	// Start returns the applied status even when only publishing it failed.
	Start(ctx context.Context) (*entity.AIAccessStatus, error)
	// Stop returns the stopped status even when publishing it failed.
	Stop(ctx context.Context) (entity.AIAccessStatus, error)
}
