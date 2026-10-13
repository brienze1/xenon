package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAppSettings struct {
	findAppSettings usecase.FindAppSettings
	saveAppSettings usecase.SaveAppSettings
	serveAIAccess   usecase.ServeAIAccess
}

func NewManageAppSettingsService(findAppSettings usecase.FindAppSettings, saveAppSettings usecase.SaveAppSettings, serveAIAccess usecase.ServeAIAccess) adapter.ManageAppSettingsService {
	return &manageAppSettings{findAppSettings: findAppSettings, saveAppSettings: saveAppSettings, serveAIAccess: serveAIAccess}
}

func (m *manageAppSettings) Get(ctx context.Context) (*entity.AppSettings, error) {
	logger.Info(ctx, "Started")

	settings, err := m.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", settings)
	return settings, nil
}

func (m *manageAppSettings) Save(ctx context.Context, patch entity.AppSettingsPatch) (*entity.AppSettings, error) {
	logger.Info(ctx, "Started", patch)

	current, err := m.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings := current.Merged(patch)
	if err := usecase.ValidateAppSettings(settings); err != nil {
		return nil, err
	}
	if err := m.saveAppSettings.SaveAppSettings(ctx, settings); err != nil {
		return nil, err
	}
	if patch.ChangesAIAccess() {
		m.serveAIAccess.Apply(ctx, settings.AIAccess)
	}

	logger.Info(ctx, "Finished", settings)
	return &settings, nil
}
