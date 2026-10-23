package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRun interface {
	SaveRun(ctx context.Context, run entity.Run) error
}
