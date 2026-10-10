package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRequestExamples interface {
	SaveExamples(ctx context.Context, root string, request entity.Request) (*entity.Request, error)
}
