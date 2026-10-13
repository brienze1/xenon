package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAIAccess struct {
	findAppSettings usecase.FindAppSettings
	serveAIAccess   usecase.ServeAIAccess
}

func NewManageAIAccessService(findAppSettings usecase.FindAppSettings, serveAIAccess usecase.ServeAIAccess) adapter.ManageAIAccessService {
	return &manageAIAccess{findAppSettings: findAppSettings, serveAIAccess: serveAIAccess}
}

func (m *manageAIAccess) Status(ctx context.Context) entity.AIAccessStatus {
	logger.Info(ctx, "Started")

	status := m.serveAIAccess.Status(ctx)

	logger.Info(ctx, "Finished", status)
	return status
}

func (m *manageAIAccess) Start(ctx context.Context) (*entity.AIAccessStatus, error) {
	logger.Info(ctx, "Started")

	settings, err := m.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	status := m.serveAIAccess.Apply(ctx, settings.AIAccess)

	logger.Info(ctx, "Finished", status)
	return &status, nil
}

func (m *manageAIAccess) Stop(ctx context.Context) entity.AIAccessStatus {
	logger.Info(ctx, "Started")

	status := m.serveAIAccess.Stop(ctx)

	logger.Info(ctx, "Finished", status)
	return status
}
