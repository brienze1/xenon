package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveImportSource interface {
	Save(ctx context.Context, source entity.ImportSource) error
}
