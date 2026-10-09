package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type editRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	appVersion            string
}

func NewEditRequestService(findWorkspaceRegistry usecase.FindWorkspaceRegistry, appVersion string) adapter.EditRequestService {
	return &editRequest{findWorkspaceRegistry: findWorkspaceRegistry, appVersion: appVersion}
}

func (e *editRequest) ApplyUrl(ctx context.Context, request entity.Request) entity.Request {
	logger.Info(ctx, "Started", request)

	result := usecase.DeriveParamsFromUrl(request)

	logger.Info(ctx, "Finished", result)
	return result
}

func (e *editRequest) ApplyParams(ctx context.Context, request entity.Request) entity.Request {
	logger.Info(ctx, "Started", request)

	result := usecase.BuildUrlFromParams(request)

	logger.Info(ctx, "Finished", result)
	return result
}

func (e *editRequest) ParseBulk(ctx context.Context, text string) []entity.KeyValue {
	logger.Info(ctx, "Started", map[string]any{"length": len(text)})

	rows := usecase.ParseBulkRows(text)

	logger.Info(ctx, "Finished", map[string]any{"rows": len(rows)})
	return rows
}

func (e *editRequest) FormatBulk(ctx context.Context, rows []entity.KeyValue) string {
	logger.Info(ctx, "Started", map[string]any{"rows": len(rows)})

	text := usecase.FormatBulkRows(rows)

	logger.Info(ctx, "Finished", map[string]any{"length": len(text)})
	return text
}

func (e *editRequest) AutoHeaders(ctx context.Context, localId string, request entity.Request) ([]entity.AutoHeader, error) {
	logger.Info(ctx, "Started", request, map[string]any{"localId": localId})

	workspace, err := e.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	effective := usecase.ResolveEffectiveRequest(request, []entity.Collection{}, workspace.Path)
	headers := usecase.ComputeAutoHeaders(effective, e.appVersion)

	logger.Info(ctx, "Finished", map[string]any{"headers": len(headers)})
	return headers, nil
}
