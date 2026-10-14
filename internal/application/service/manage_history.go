package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageHistory struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findAppSettings       usecase.FindAppSettings
	findHistory           usecase.FindHistory
	clearHistory          usecase.ClearHistory
	publishAIEvent        usecase.PublishAIEvent
}

func NewManageHistoryService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findAppSettings usecase.FindAppSettings,
	findHistory usecase.FindHistory,
	clearHistory usecase.ClearHistory,
	publishAIEvent usecase.PublishAIEvent,
) adapter.ManageHistoryService {
	return &manageHistory{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findAppSettings:       findAppSettings,
		findHistory:           findHistory,
		clearHistory:          clearHistory,
		publishAIEvent:        publishAIEvent,
	}
}

func (h *manageHistory) List(ctx context.Context, localId string, query entity.HistoryQuery) ([]entity.HistoryEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "query": query})

	if _, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	settings, err := h.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	stored, err := h.findHistory.Find(ctx, localId, settings.HistoryCaps())
	if err != nil {
		return nil, err
	}
	page := stored.Page(query)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "query": query, "result": entity.HistoryEntries(page)})
	return page, nil
}

func (h *manageHistory) Get(ctx context.Context, localId, id string) (*entity.HistoryEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	if _, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	settings, err := h.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := h.findHistory.FindEntry(ctx, localId, id, settings.HistoryCaps())
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *entry})
	return entry, nil
}

func (h *manageHistory) Clear(ctx context.Context, localId string, source enums.HistoryFilter) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "source": source})

	if _, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := h.clearHistory.Clear(ctx, localId, source); err != nil {
		return err
	}
	if err := h.publishAIEvent.HistoryChanged(ctx, localId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "source": source})
	return nil
}
