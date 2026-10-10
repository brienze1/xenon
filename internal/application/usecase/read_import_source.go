package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadImportSource interface {
	Read(ctx context.Context, request entity.ImportRequest) ([]entity.ImportInput, []string, error)
}
