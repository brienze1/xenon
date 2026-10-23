package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PickPathService interface {
	PickFolder(ctx context.Context, title string) (string, error)
	PickFile(ctx context.Context, title string, filters []entity.FileFilter) (string, error)
	PickSaveFile(ctx context.Context, title, suggestedName string) (string, error)
}
