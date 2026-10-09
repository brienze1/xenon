package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageHistory struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findHistory           usecase.FindHistory
	clearHistory          usecase.ClearHistory
}

func NewManageHistoryService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findHistory usecase.FindHistory,
	clearHistory usecase.ClearHistory,
) adapter.ManageHistoryService {
	return &manageHistory{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findHistory:           findHistory,
		clearHistory:          clearHistory,
	}
}

func (h *manageHistory) List(ctx context.Context, localId string, query entity.HistoryQuery) ([]entity.HistoryEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "query": query})

	if _, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	stored, err := h.findHistory.Find(ctx, localId)
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
	entry, err := h.findHistory.FindEntry(ctx, localId, id)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *entry})
	return entry, nil
}

func (h *manageHistory) Clear(ctx context.Context, localId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	if _, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := h.clearHistory.Clear(ctx, localId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}
