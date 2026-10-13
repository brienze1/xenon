package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindKnownHosts interface {
	FindKnownHosts(ctx context.Context) (*entity.KnownHosts, error)
}
