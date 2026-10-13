package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindConsoleEntry interface {
	FindEntry(ctx context.Context, id string) (*entity.ConsoleEntry, error)
}
