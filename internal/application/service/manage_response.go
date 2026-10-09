package service

import (
	"context"
	"io"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageResponse struct {
	findResponse       usecase.FindResponse
	releaseResponse    usecase.ReleaseResponse
	exportResponseBody usecase.ExportResponseBody
}

func NewManageResponseService(
	findResponse usecase.FindResponse,
	releaseResponse usecase.ReleaseResponse,
	exportResponseBody usecase.ExportResponseBody,
) adapter.ManageResponseService {
	return &manageResponse{
		findResponse:       findResponse,
		releaseResponse:    releaseResponse,
		exportResponseBody: exportResponseBody,
	}
}

func (m *manageResponse) LastResult(ctx context.Context, tabId string) (*entity.Response, error) {
	logger.Info(ctx, "Started", map[string]any{"tabId": tabId})

	response, err := m.findResponse.Find(ctx, tabId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", response, map[string]any{"tabId": tabId})
	return response, nil
}

func (m *manageResponse) OpenBody(ctx context.Context, tabId string) (*entity.Response, io.ReadCloser, error) {
	logger.Info(ctx, "Started", map[string]any{"tabId": tabId})

	response, err := m.findResponse.Find(ctx, tabId)
	if err != nil {
		return nil, nil, err
	}
	body, err := m.findResponse.OpenBody(ctx, tabId)
	if err != nil {
		return nil, nil, err
	}

	logger.Info(ctx, "Finished", response, map[string]any{"tabId": tabId})
	return response, body, nil
}

func (m *manageResponse) SaveBody(ctx context.Context, tabId, path string) error {
	logger.Info(ctx, "Started", map[string]any{"tabId": tabId, "path": path})

	if err := m.exportResponseBody.Export(ctx, tabId, path); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"tabId": tabId, "path": path})
	return nil
}

func (m *manageResponse) Release(ctx context.Context, tabId string) error {
	logger.Info(ctx, "Started", map[string]any{"tabId": tabId})

	if err := m.releaseResponse.Release(ctx, tabId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"tabId": tabId})
	return nil
}

func (m *manageResponse) ReleaseAll(ctx context.Context) error {
	logger.Info(ctx, "Started")

	if err := m.releaseResponse.ReleaseAll(ctx); err != nil {
		return err
	}

	logger.Info(ctx, "Finished")
	return nil
}
