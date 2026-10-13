package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadPostmanEnvironment interface {
	ReadEnvironment(ctx context.Context, document entity.ImportDocument) (*entity.ImportVariables, error)
}
