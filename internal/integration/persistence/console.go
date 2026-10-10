package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type consolePersistence struct {
	buffer entity.ConsoleBuffer
	mutex  sync.Mutex
}

func NewConsolePersistence() adapter.ConsolePersistence {
	return &consolePersistence{buffer: entity.NewConsoleBuffer()}
}

func (c *consolePersistence) Append(ctx context.Context, entries []entity.ConsoleEntry) {
	logger.Debug(ctx, "Started", map[string]any{"count": len(entries)})

	c.mutex.Lock()
	c.buffer = c.buffer.Appended(entries)
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"count": len(entries)})
}

func (c *consolePersistence) FindEntries(ctx context.Context, afterId string, limit int) []entity.ConsoleEntry {
	logger.Debug(ctx, "Started", map[string]any{"afterId": afterId, "limit": limit})

	c.mutex.Lock()
	entries := c.buffer.After(afterId, limit)
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"afterId": afterId, "limit": limit, "count": len(entries)})
	return entries
}

func (c *consolePersistence) Clear(ctx context.Context) {
	logger.Debug(ctx, "Started")

	c.mutex.Lock()
	c.buffer = entity.NewConsoleBuffer()
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished")
}
