package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type uiState struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findUIState           usecase.FindUIState
	saveUIState           usecase.SaveUIState
}

func NewUIStateService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findUIState usecase.FindUIState,
	saveUIState usecase.SaveUIState,
) adapter.UIStateService {
	return &uiState{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findUIState:           findUIState,
		saveUIState:           saveUIState,
	}
}

func (u *uiState) Get(ctx context.Context, localId string) (*entity.UIState, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	if _, err := u.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	state, err := u.findUIState.Find(ctx, localId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return state, nil
}

func (u *uiState) Save(ctx context.Context, localId string, state entity.UIState) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "tabs": len(state.Tabs)})

	if _, err := u.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := u.saveUIState.Save(ctx, localId, state.Normalized()); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return nil
}
