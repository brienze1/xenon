package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAppSettings struct {
	findAppSettings                usecase.FindAppSettings
	saveAppSettings                usecase.SaveAppSettings
	serveAIAccess                  usecase.ServeAIAccess
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus
	publishAIEvent                 usecase.PublishAIEvent
}

func NewManageAppSettingsService(
	findAppSettings usecase.FindAppSettings,
	saveAppSettings usecase.SaveAppSettings,
	serveAIAccess usecase.ServeAIAccess,
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus,
	publishAIEvent usecase.PublishAIEvent,
) adapter.ManageAppSettingsService {
	return &manageAppSettings{
		findAppSettings:                findAppSettings,
		saveAppSettings:                saveAppSettings,
		serveAIAccess:                  serveAIAccess,
		replacePublishedAIAccessStatus: replacePublishedAIAccessStatus,
		publishAIEvent:                 publishAIEvent,
	}
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
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	if err := m.saveAppSettings.SaveAppSettings(ctx, settings); err != nil {
		return nil, err
	}
	if patch.ChangesAIAccess() {
		if err := m.publish(ctx, m.serve(ctx, settings.AIAccess)); err != nil {
			return &settings, err
		}
	}

	logger.Info(ctx, "Finished", patch, settings)
	return &settings, nil
}

func (m *manageAppSettings) serve(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	if settings.ShouldListen() {
		return m.serveAIAccess.Apply(ctx, settings)
	}
	return m.serveAIAccess.Disable(ctx, settings)
}

func (m *manageAppSettings) publish(ctx context.Context, status entity.AIAccessStatus) error {
	if !status.Differs(m.replacePublishedAIAccessStatus.ReplacePublished(ctx, status)) {
		return nil
	}
	return m.publishAIEvent.AccessStatusChanged(ctx, status)
}
