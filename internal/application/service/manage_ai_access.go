package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAIAccess struct {
	findAppSettings                usecase.FindAppSettings
	serveAIAccess                  usecase.ServeAIAccess
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus
	publishAIEvent                 usecase.PublishAIEvent
}

func NewManageAIAccessService(
	findAppSettings usecase.FindAppSettings,
	serveAIAccess usecase.ServeAIAccess,
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus,
	publishAIEvent usecase.PublishAIEvent,
) adapter.ManageAIAccessService {
	return &manageAIAccess{
		findAppSettings:                findAppSettings,
		serveAIAccess:                  serveAIAccess,
		replacePublishedAIAccessStatus: replacePublishedAIAccessStatus,
		publishAIEvent:                 publishAIEvent,
	}
}

func (m *manageAIAccess) Start(ctx context.Context) (*entity.AIAccessStatus, error) {
	logger.Info(ctx, "Started")

	settings, err := m.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	status := m.serve(ctx, settings.AIAccess)
	if err := m.publish(ctx, status); err != nil {
		return &status, err
	}

	logger.Info(ctx, "Finished", status)
	return &status, nil
}

func (m *manageAIAccess) Stop(ctx context.Context) (entity.AIAccessStatus, error) {
	logger.Info(ctx, "Started")

	status := m.serveAIAccess.Stop(ctx)
	if err := m.publish(ctx, status); err != nil {
		return status, err
	}

	logger.Info(ctx, "Finished", status)
	return status, nil
}

func (m *manageAIAccess) serve(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	if settings.ShouldListen() {
		return m.serveAIAccess.Apply(ctx, settings)
	}
	return m.serveAIAccess.Disable(ctx, settings)
}

func (m *manageAIAccess) publish(ctx context.Context, status entity.AIAccessStatus) error {
	if !status.Differs(m.replacePublishedAIAccessStatus.ReplacePublished(ctx, status)) {
		return nil
	}
	return m.publishAIEvent.AccessStatusChanged(ctx, status)
}
