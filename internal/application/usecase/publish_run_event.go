package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishRunEvent interface {
	Progress(ctx context.Context, run entity.Run) error
	Finished(ctx context.Context, run entity.Run) error
}
