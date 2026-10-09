package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type InspectFileChange interface {
	Inspect(ctx context.Context, root, path string) (*entity.FileChange, error)
}
