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
	findTreeItem          usecase.FindTreeItem
	readCurlCommand       usecase.ReadCurlCommand
}

func NewEditRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	readCurlCommand usecase.ReadCurlCommand,
) adapter.EditRequestService {
	return &editRequest{findWorkspaceRegistry: findWorkspaceRegistry, findTreeItem: findTreeItem, readCurlCommand: readCurlCommand}
}

func (e *editRequest) ApplyUrl(ctx context.Context, request entity.Request) entity.Request {
	logger.Info(ctx, "Started", map[string]any{"request": request})

	result := usecase.DeriveParamsFromUrl(request)

	logger.Info(ctx, "Finished", map[string]any{"request": request, "result": result})
	return result
}

func (e *editRequest) ApplyParams(ctx context.Context, request entity.Request) entity.Request {
	logger.Info(ctx, "Started", map[string]any{"request": request})

	result := usecase.BuildUrlFromParams(request)

	logger.Info(ctx, "Finished", map[string]any{"request": request, "result": result})
	return result
}

func (e *editRequest) ParseBulk(ctx context.Context, text string) []entity.KeyValue {
	logger.Info(ctx, "Started", map[string]any{"textLength": len(text)})

	rows := usecase.ParseBulkRows(text)

	logger.Info(ctx, "Finished", map[string]any{"textLength": len(text), "result": entity.KeyValues(rows)})
	return rows
}

func (e *editRequest) ParseCurl(ctx context.Context, text string) (*entity.ParsedCurl, error) {
	logger.Info(ctx, "Started", map[string]any{"textLength": len(text)})

	parsed, err := e.readCurlCommand.ReadCommand(ctx, text)
	if err != nil {
		return nil, err
	}
	parsed.Request = usecase.DeriveParamsFromUrl(parsed.Request)

	logger.Info(ctx, "Finished", map[string]any{"textLength": len(text), "result": *parsed})
	return parsed, nil
}

func (e *editRequest) FormatBulk(ctx context.Context, rows []entity.KeyValue) string {
	logger.Info(ctx, "Started", map[string]any{"rows": entity.KeyValues(rows)})

	text := usecase.FormatBulkRows(rows)

	logger.Info(ctx, "Finished", map[string]any{"rows": entity.KeyValues(rows), "resultLength": len(text)})
	return text
}

func (e *editRequest) AutoHeaders(ctx context.Context, localId string, request entity.Request, appVersion string) ([]entity.AutoHeader, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "request": request, "appVersion": appVersion})

	workspace, err := e.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	ancestors := make([]entity.Collection, 0)
	if request.HasID() {
		ancestors, err = e.findTreeItem.FindAncestors(ctx, workspace.Path, request.ID)
		if err != nil {
			return nil, err
		}
	}
	effective := usecase.ResolveEffectiveRequest(request, ancestors, workspace.Path)
	headers := usecase.ComputeAutoHeaders(effective, appVersion)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "request": request, "appVersion": appVersion, "result": entity.AutoHeaders(headers)})
	return headers, nil
}
