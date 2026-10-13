package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const consoleEntryNotKept = "the console no longer keeps this entry"

type consolePersistence struct {
	buffer entity.ConsoleBuffer
	mutex  sync.Mutex
}

func NewConsolePersistence() adapter.ConsolePersistence {
	return &consolePersistence{buffer: entity.NewConsoleBuffer()}
}

func (c *consolePersistence) Append(ctx context.Context, entries []entity.ConsoleEntry) []entity.ConsoleEntry {
	logger.Debug(ctx, "Started", map[string]any{"count": len(entries)})

	c.mutex.Lock()
	buffer, dropped := c.buffer.Appended(entries)
	c.buffer = buffer
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"count": len(entries), "dropped": len(dropped)})
	return dropped
}

func (c *consolePersistence) FindEntries(ctx context.Context, query entity.ConsoleQuery) []entity.ConsoleEntry {
	logger.Debug(ctx, "Started", map[string]any{"query": query})

	c.mutex.Lock()
	entries := c.buffer.Find(query)
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"query": query, "count": len(entries)})
	return entries
}

func (c *consolePersistence) FindEntry(ctx context.Context, id string) (*entity.ConsoleEntry, error) {
	logger.Debug(ctx, "Started", map[string]any{"id": id})

	c.mutex.Lock()
	entry, found := c.buffer.Entry(id)
	c.mutex.Unlock()
	if !found {
		return nil, errs.NotFoundError(consoleEntryNotKept, "CON-01404")
	}

	logger.Debug(ctx, "Finished", map[string]any{"id": id})
	return &entry, nil
}

func (c *consolePersistence) Clear(ctx context.Context) []entity.ConsoleEntry {
	logger.Debug(ctx, "Started")

	c.mutex.Lock()
	removed := c.buffer.Entries
	c.buffer = entity.NewConsoleBuffer()
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"count": len(removed)})
	return removed
}
