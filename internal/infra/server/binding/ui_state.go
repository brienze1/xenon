package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type UIStateService struct {
	uiState adapter.UIStateController
}

func NewUIStateService(uiState adapter.UIStateController) *UIStateService {
	return &UIStateService{uiState: uiState}
}

func (u *UIStateService) Get(ctx context.Context, request dto.GetUIState) (*dto.UIState, error) {
	return u.uiState.Get(ctx, request)
}

func (u *UIStateService) Save(ctx context.Context, request dto.UIState) error {
	return u.uiState.Save(ctx, request)
}
