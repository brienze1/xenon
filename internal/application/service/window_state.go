package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type windowState struct {
	findWindowState usecase.FindWindowState
	saveWindowState usecase.SaveWindowState
}

func NewWindowStateService(findWindowState usecase.FindWindowState, saveWindowState usecase.SaveWindowState) adapter.WindowStateService {
	return &windowState{findWindowState: findWindowState, saveWindowState: saveWindowState}
}

func (w *windowState) Restore(ctx context.Context, screens []entity.Screen) (*entity.WindowState, error) {
	logger.Info(ctx, "Started", map[string]any{"screens": screens})

	saved, err := w.findWindowState.FindWindow(ctx)
	if err != nil {
		return nil, err
	}
	fitted := saved.FittedTo(screens)

	logger.Info(ctx, "Finished", fitted)
	return &fitted, nil
}

func (w *windowState) Save(ctx context.Context, state entity.WindowState) error {
	logger.Info(ctx, "Started", state)

	saved, err := w.findWindowState.FindWindow(ctx)
	if err != nil {
		return err
	}
	if err := w.saveWindowState.SaveWindow(ctx, saved.Updated(state)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", state)
	return nil
}
