package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type DialogController interface {
	PickFolder(ctx context.Context, request dto.PickFolder) (*dto.PickedPath, error)
	PickFile(ctx context.Context, request dto.PickFile) (*dto.PickedPath, error)
	PickSaveFile(ctx context.Context, request dto.PickSaveFile) (*dto.PickedPath, error)
}
