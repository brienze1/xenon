package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WindowService struct {
	windowState adapter.WindowStateController
}

func NewWindowService(windowState adapter.WindowStateController) *WindowService {
	return &WindowService{windowState: windowState}
}

func (w *WindowService) Restore(ctx context.Context, request dto.RestoreWindow) (*dto.WindowState, error) {
	return w.windowState.Restore(ctx, request)
}

func (w *WindowService) Save(ctx context.Context, request dto.WindowState) error {
	return w.windowState.Save(ctx, request)
}
