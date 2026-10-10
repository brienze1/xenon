package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type AppendConsoleEntries interface {
	Append(ctx context.Context, entries []entity.ConsoleEntry)
}
