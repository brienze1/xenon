package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

const stoppedRunWait = 5 * time.Second

func stopActiveRuns(ctx context.Context, findActiveRunControls usecase.FindActiveRunControls, localId string) {
	active := findActiveRunControls.FindControlsByWorkspace(ctx, localId)
	for _, run := range active {
		run.Control.Stop()
	}
	awaitRuns(ctx, active)
}

func awaitRuns(ctx context.Context, active []usecase.ActiveRun) {
	deadline, cancel := context.WithTimeout(context.WithoutCancel(ctx), stoppedRunWait)
	defer cancel()
	for _, run := range active {
		if !run.Control.AwaitFinished(deadline) {
			logger.Info(ctx, "A stopped run is still finishing", map[string]any{"localId": run.Run.Target.LocalID, "runId": run.Run.ID})
			return
		}
	}
}
