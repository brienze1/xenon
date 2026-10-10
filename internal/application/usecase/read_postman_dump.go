package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadPostmanDump interface {
	ReadDump(ctx context.Context, input entity.ImportInput) (*entity.PostmanDump, error)
}
