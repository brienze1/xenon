package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WindowStateController interface {
	Restore(ctx context.Context, request dto.RestoreWindow) (*dto.WindowState, error)
	Save(ctx context.Context, request dto.WindowState) error
}
