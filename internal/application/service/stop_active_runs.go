package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

const stoppedRunWait = 5 * time.Second

func stopActiveRuns(ctx context.Context, manageActiveRuns usecase.ManageActiveRuns, localId string) {
	active := manageActiveRuns.ListWorkspace(ctx, localId)
	for _, run := range active {
		run.Control.Stop()
	}
	deadline := time.After(stoppedRunWait)
	for _, run := range active {
		select {
		case <-run.Control.Done():
		case <-deadline:
			logger.Info(ctx, "A stopped run is still finishing", map[string]any{"localId": localId, "runId": run.Run.ID})
			return
		}
	}
}
