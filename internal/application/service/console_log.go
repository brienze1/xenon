package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

// consoleLog keeps the session console: entries are stored, their large bodies kept for the body route, and the UI told.
type consoleLog struct {
	appendConsoleEntries  usecase.AppendConsoleEntries
	clearConsole          usecase.ClearConsole
	publishConsoleEntries usecase.PublishConsoleEntries
	saveResponse          usecase.SaveResponse
	releaseResponse       usecase.ReleaseResponse
}

func (c consoleLog) append(ctx context.Context, entries ...entity.ConsoleEntry) {
	if len(entries) == 0 {
		return
	}
	kept := make([]entity.ConsoleEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.ServesBodyByRoute() {
			if err := c.saveResponse.Save(ctx, entry.BodyKey(), *entry.Response); err != nil {
				logger.Warn(ctx, err, "The body of a console entry could not be kept", map[string]any{"id": entry.ID})
			}
		}
		kept = append(kept, entry.WithoutRoutedBody())
	}
	c.release(ctx, c.appendConsoleEntries.Append(ctx, kept))
	c.publishConsoleEntries.PublishEntries(ctx, kept)
}

func (c consoleLog) clear(ctx context.Context) {
	c.release(ctx, c.clearConsole.Clear(ctx))
	c.publishConsoleEntries.PublishCleared(ctx)
}

func (c consoleLog) release(ctx context.Context, entries []entity.ConsoleEntry) {
	for _, entry := range entries {
		if !entry.ServesBodyByRoute() {
			continue
		}
		if err := c.releaseResponse.Release(ctx, entry.BodyKey()); err != nil {
			logger.Warn(ctx, err, "The body of a dropped console entry could not be released", map[string]any{"id": entry.ID})
		}
	}
}
