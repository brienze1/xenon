package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type saveRunnerPreferences struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveRunnerPreferences usecase.SaveRunnerPreferences
}

func NewSaveRunnerPreferencesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveStoredRunnerPreferences usecase.SaveRunnerPreferences,
) adapter.SaveRunnerPreferencesService {
	return &saveRunnerPreferences{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveRunnerPreferences: saveStoredRunnerPreferences,
	}
}

func (s *saveRunnerPreferences) Save(ctx context.Context, localId, itemId string, preferences entity.RunnerPreferences) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId, "preferences": preferences})

	if _, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := s.saveRunnerPreferences.SaveRunnerPreferences(ctx, localId, itemId, preferences); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "preferences": preferences})
	return nil
}
