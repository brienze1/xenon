package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageConsole struct {
	findConsoleEntries usecase.FindConsoleEntries
	clearConsole       usecase.ClearConsole
}

func NewManageConsoleService(findConsoleEntries usecase.FindConsoleEntries, clearConsole usecase.ClearConsole) adapter.ManageConsoleService {
	return &manageConsole{findConsoleEntries: findConsoleEntries, clearConsole: clearConsole}
}

func (m *manageConsole) List(ctx context.Context, afterId string, limit int) []entity.ConsoleEntry {
	logger.Info(ctx, "Started", map[string]any{"afterId": afterId, "limit": limit})

	entries := m.findConsoleEntries.FindEntries(ctx, afterId, limit)

	logger.Info(ctx, "Finished", map[string]any{"afterId": afterId, "limit": limit, "count": len(entries)})
	return entries
}

func (m *manageConsole) Clear(ctx context.Context) {
	logger.Info(ctx, "Started")

	m.clearConsole.Clear(ctx)

	logger.Info(ctx, "Finished")
}
