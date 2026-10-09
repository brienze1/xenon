package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageWindowState struct {
	findWindowState usecase.FindWindowState
	saveWindowState usecase.SaveWindowState
}

func NewManageWindowStateService(findWindowState usecase.FindWindowState, saveWindowState usecase.SaveWindowState) adapter.ManageWindowStateService {
	return &manageWindowState{findWindowState: findWindowState, saveWindowState: saveWindowState}
}

func (w *manageWindowState) Restore(ctx context.Context, screens []entity.Screen) (*entity.WindowState, error) {
	logger.Info(ctx, "Started", map[string]any{"screens": screens})

	saved, err := w.findWindowState.FindWindow(ctx)
	if err != nil {
		return nil, err
	}
	fitted := saved.FittedTo(screens)

	logger.Info(ctx, "Finished", map[string]any{"screens": screens, "result": fitted})
	return &fitted, nil
}

func (w *manageWindowState) Save(ctx context.Context, state entity.WindowState) error {
	logger.Info(ctx, "Started", map[string]any{"state": state})

	if !state.IsPersistable() {
		logger.Info(ctx, "Finished", map[string]any{"state": state})
		return nil
	}
	saved, err := w.findWindowState.FindWindow(ctx)
	if err != nil {
		return err
	}
	if err := w.saveWindowState.SaveWindow(ctx, saved.Updated(state)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"state": state})
	return nil
}
