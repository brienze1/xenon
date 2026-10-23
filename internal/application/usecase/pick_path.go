package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PickPath interface {
	PickFolder(ctx context.Context, title string) (string, error)
	PickFile(ctx context.Context, title string, filters []entity.FileFilter) (string, error)
	PickSaveFile(ctx context.Context, title, suggestedName string) (string, error)
	PickSources(ctx context.Context, title string) ([]string, error)
}
