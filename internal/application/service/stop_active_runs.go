package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

const stoppedRunWait = 5 * time.Second

func stopActiveRuns(ctx context.Context, manageActiveRuns usecase.ManageActiveRuns, localId string) {
	active := manageActiveRuns.ListWorkspace(ctx, localId)
	for _, run := range active {
		run.Control.Stop()
	}
	awaitRuns(ctx, active)
}

func awaitRuns(ctx context.Context, active []entity.ActiveRun) {
	deadline := time.After(stoppedRunWait)
	for _, run := range active {
		select {
		case <-run.Control.Done():
		case <-deadline:
			logger.Info(ctx, "A stopped run is still finishing", map[string]any{"localId": run.Run.Target.LocalID, "runId": run.Run.ID})
			return
		}
	}
}
