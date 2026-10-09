package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type UIStateService struct {
	uiState adapter.UIStateController
}

func NewUIStateService() *UIStateService {
	return &UIStateService{uiState: controller.GetUIStateController()}
}

func (u *UIStateService) Get(ctx context.Context, request dto.GetUIState) (*dto.UIState, error) {
	return u.uiState.Get(ctx, request)
}

func (u *UIStateService) Save(ctx context.Context, request dto.UIState) error {
	return u.uiState.Save(ctx, request)
}
