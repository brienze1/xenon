package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

type deleteRuns struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findActiveRuns        usecase.FindActiveRuns
	deleteRuns            usecase.DeleteRuns
	releaseResponse       usecase.ReleaseResponse
}

func NewDeleteRunsService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findActiveRuns usecase.FindActiveRuns,
	deleteStoredRuns usecase.DeleteRuns,
	releaseResponse usecase.ReleaseResponse,
) adapter.DeleteRunsService {
	return &deleteRuns{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findActiveRuns:        findActiveRuns,
		deleteRuns:            deleteStoredRuns,
		releaseResponse:       releaseResponse,
	}
}

func (d *deleteRuns) Delete(ctx context.Context, localId, runId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	if active, found := d.findActiveRuns.FindById(ctx, localId, runId); found {
		return active.ActiveConflict()
	}
	deleted, err := d.deleteRuns.DeleteRun(ctx, localId, runId)
	if err != nil {
		return err
	}
	if err := releaseRunResponses(ctx, d.releaseResponse, *deleted); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
	return nil
}

func (d *deleteRuns) Clear(ctx context.Context, localId, itemId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	if _, err := d.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	deleted, err := d.deleteRuns.ClearRuns(ctx, localId, itemId)
	if err != nil {
		return err
	}
	if err := releaseRunResponses(ctx, d.releaseResponse, deleted...); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "runs": len(deleted)})
	return nil
}
