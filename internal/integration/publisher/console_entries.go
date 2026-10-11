package publisher

import (
	"context"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	consoleBatchWindow  = 50 * time.Millisecond
	consoleBatchEntries = 100
)

// consoleEntriesPublisher batches entries into one console:entries event at most every 50 ms or every 100 entries.
type consoleEntriesPublisher struct {
	eventEmitter adapter.EventEmitter
	timer        *time.Timer
	pending      []entity.ConsoleEntry
	mutex        sync.Mutex
}

func NewConsoleEntries(eventEmitter adapter.EventEmitter) adapter.ConsoleEntriesPublisher {
	return &consoleEntriesPublisher{eventEmitter: eventEmitter}
}

func (c *consoleEntriesPublisher) PublishEntries(ctx context.Context, entries []entity.ConsoleEntry) {
	logger.Debug(ctx, "Started", map[string]any{"count": len(entries)})

	c.mutex.Lock()
	c.pending = append(c.pending, entries...)
	switch {
	case len(c.pending) >= consoleBatchEntries:
		c.flushLocked()
	case c.timer == nil && len(c.pending) > 0:
		c.timer = time.AfterFunc(consoleBatchWindow, c.flush)
	}
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"count": len(entries)})
}

func (c *consoleEntriesPublisher) PublishCleared(ctx context.Context) {
	logger.Debug(ctx, "Started")

	c.mutex.Lock()
	c.stopTimer()
	c.pending = nil
	c.eventEmitter.Emit(dto.ConsoleEntriesEvent, dto.NewConsoleCleared())
	c.mutex.Unlock()

	logger.Debug(ctx, "Finished")
}

func (c *consoleEntriesPublisher) flush() {
	c.mutex.Lock()
	c.flushLocked()
	c.mutex.Unlock()
}

func (c *consoleEntriesPublisher) flushLocked() {
	c.stopTimer()
	if len(c.pending) == 0 {
		return
	}
	c.eventEmitter.Emit(dto.ConsoleEntriesEvent, dto.NewConsoleEntries(c.pending))
	c.pending = nil
}

func (c *consoleEntriesPublisher) stopTimer() {
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}
