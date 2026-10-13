package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishConsoleEntries interface {
	PublishEntries(ctx context.Context, entries []entity.ConsoleEntry)
	PublishCleared(ctx context.Context)
}
