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
	findConsoleEntry   usecase.FindConsoleEntry
	console            consoleLog
}

func NewManageConsoleService(
	findConsoleEntries usecase.FindConsoleEntries,
	findConsoleEntry usecase.FindConsoleEntry,
	appendConsoleEntries usecase.AppendConsoleEntries,
	clearConsole usecase.ClearConsole,
	publishConsoleEntries usecase.PublishConsoleEntries,
	saveResponse usecase.SaveResponse,
	releaseResponse usecase.ReleaseResponse,
) adapter.ManageConsoleService {
	return &manageConsole{
		findConsoleEntries: findConsoleEntries,
		findConsoleEntry:   findConsoleEntry,
		console: consoleLog{
			appendConsoleEntries:  appendConsoleEntries,
			clearConsole:          clearConsole,
			publishConsoleEntries: publishConsoleEntries,
			saveResponse:          saveResponse,
			releaseResponse:       releaseResponse,
		},
	}
}

func (m *manageConsole) List(ctx context.Context, query entity.ConsoleQuery) []entity.ConsoleEntry {
	logger.Info(ctx, "Started", map[string]any{"query": query})

	entries := m.findConsoleEntries.FindEntries(ctx, query)

	logger.Info(ctx, "Finished", map[string]any{"query": query, "count": len(entries)})
	return entries
}

func (m *manageConsole) Get(ctx context.Context, id string) (*entity.ConsoleEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"id": id})

	entry, err := m.findConsoleEntry.FindEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"id": id, "kind": entry.Kind})
	return entry, nil
}

func (m *manageConsole) Clear(ctx context.Context) {
	logger.Info(ctx, "Started")

	m.console.clear(ctx)

	logger.Info(ctx, "Finished")
}
