package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageUIState struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findUIState           usecase.FindUIState
	saveUIState           usecase.SaveUIState
}

func NewManageUIStateService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findUIState usecase.FindUIState,
	saveUIState usecase.SaveUIState,
) adapter.ManageUIStateService {
	return &manageUIState{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findUIState:           findUIState,
		saveUIState:           saveUIState,
	}
}

func (u *manageUIState) Get(ctx context.Context, localId string) (*entity.UIState, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	if _, err := u.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	state, err := u.findUIState.Find(ctx, localId)
	if state, err = uiStateOrDefaults(ctx, localId, state, err); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": *state})
	return state, nil
}

func (u *manageUIState) Save(ctx context.Context, localId string, state entity.UIState) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "state": state})

	if _, err := u.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := u.saveUIState.Save(ctx, localId, state.Normalized()); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "state": state})
	return nil
}

func uiStateOrDefaults(ctx context.Context, localId string, state *entity.UIState, err error) (*entity.UIState, error) {
	switch {
	case err == nil:
		return state, nil
	case errs.IsNotFoundError(err):
		logger.Info(ctx, "UI state not saved yet, using the defaults", map[string]any{"localId": localId})
	case errs.IsUnprocessableEntityError(err):
		logger.Warn(ctx, err, "state.json is not valid JSON, using the defaults", map[string]any{"localId": localId})
	default:
		return nil, err
	}
	defaults := entity.NewUIState()
	return &defaults, nil
}
