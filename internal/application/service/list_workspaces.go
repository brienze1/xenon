package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type listWorkspaces struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
}

func NewListWorkspacesService(findWorkspaceRegistry usecase.FindWorkspaceRegistry) adapter.ListWorkspacesService {
	return &listWorkspaces{findWorkspaceRegistry: findWorkspaceRegistry}
}

func (l *listWorkspaces) List(ctx context.Context) (*entity.AppRegistry, error) {
	logger.Info(ctx, "Started")

	registry, err := l.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", registry)
	return registry, nil
}
