package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type controlRun struct {
	findActiveRunControls usecase.FindActiveRunControls
	changeActiveRun       usecase.ChangeActiveRun
	publishRunEvent       usecase.PublishRunEvent
	lookup                runLookup
}

func NewControlRunService(
	findActiveRuns usecase.FindActiveRuns,
	findActiveRunControls usecase.FindActiveRunControls,
	changeActiveRun usecase.ChangeActiveRun,
	publishRunEvent usecase.PublishRunEvent,
	findRuns usecase.FindRuns,
	findResponse usecase.FindResponse,
) adapter.ControlRunService {
	return &controlRun{
		findActiveRunControls: findActiveRunControls,
		changeActiveRun:       changeActiveRun,
		publishRunEvent:       publishRunEvent,
		lookup:                runLookup{findActiveRuns: findActiveRuns, findRuns: findRuns, findResponse: findResponse},
	}
}

func (c *controlRun) Pause(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := c.steer(ctx, localId, runId, func(run *entity.Run) bool { return run.Pause() }, usecase.RunControl.Pause)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (c *controlRun) Resume(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := c.steer(ctx, localId, runId, func(run *entity.Run) bool { return run.Resume() }, usecase.RunControl.Resume)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (c *controlRun) Stop(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := c.stop(ctx, localId, runId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (c *controlRun) Abandon(ctx context.Context) {
	logger.Info(ctx, "Started")

	active := c.findActiveRunControls.FindAllControls(ctx)
	for _, run := range active {
		run.Control.Abandon()
	}
	awaitRuns(ctx, active)

	logger.Info(ctx, "Finished", map[string]any{"runs": len(active)})
}

func (c *controlRun) stop(ctx context.Context, localId, runId string) (*entity.Run, error) {
	active, err := c.findActiveRunControls.FindControlById(ctx, localId, runId)
	if errs.IsNotFoundError(err) {
		return c.lookup.held(ctx, localId, runId)
	}
	if err != nil {
		return nil, err
	}
	active.Control.Stop()
	return &active.Run, nil
}

func (c *controlRun) steer(ctx context.Context, localId, runId string, change func(*entity.Run) bool, steer func(usecase.RunControl)) (*entity.Run, error) {
	active, err := c.findActiveRunControls.FindControlById(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	if err := active.Run.ValidateSteering(); err != nil {
		return nil, err
	}
	changed := false
	steered, found := c.changeActiveRun.Change(ctx, localId, runId, func(run *entity.Run) {
		if changed = change(run); changed {
			steer(active.Control)
		}
	})
	if !found {
		return c.lookup.held(ctx, localId, runId)
	}
	if changed {
		if err := c.publishRunEvent.Progress(ctx, *steered); err != nil {
			return nil, err
		}
	}
	return steered, nil
}
