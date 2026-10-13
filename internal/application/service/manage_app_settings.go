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
}

func NewManageAppSettingsService(findAppSettings usecase.FindAppSettings, saveAppSettings usecase.SaveAppSettings) adapter.ManageAppSettingsService {
	return &manageAppSettings{findAppSettings: findAppSettings, saveAppSettings: saveAppSettings}
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

func (m *manageAppSettings) Save(ctx context.Context, settings entity.AppSettings) (*entity.AppSettings, error) {
	logger.Info(ctx, "Started", settings)

	if err := m.saveAppSettings.SaveAppSettings(ctx, settings); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", settings)
	return &settings, nil
}
