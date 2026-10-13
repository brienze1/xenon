package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindConsoleEntries interface {
	FindEntries(ctx context.Context, query entity.ConsoleQuery) []entity.ConsoleEntry
}
