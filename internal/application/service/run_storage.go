package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runexportformat"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func (m *manageRunner) List(ctx context.Context, localId, itemId string) ([]entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	runs, err := m.findRuns.FindRuns(ctx, localId, itemId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "runs": len(runs)})
	return runs, nil
}

func (m *manageRunner) Delete(ctx context.Context, localId, runId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	if _, active := m.manageActiveRuns.Find(ctx, localId, runId); active {
		return usecase.RejectActiveRun(runId)
	}
	deleted, err := m.deleteRuns.DeleteRun(ctx, localId, runId)
	if err != nil {
		return err
	}
	m.releaseResponses(ctx, *deleted)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
	return nil
}

func (m *manageRunner) Clear(ctx context.Context, localId, itemId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	deleted, err := m.deleteRuns.ClearRuns(ctx, localId, itemId)
	if err != nil {
		return err
	}
	m.releaseResponses(ctx, deleted...)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "runs": len(deleted)})
	return nil
}

func (m *manageRunner) Response(ctx context.Context, localId, runId string, index int) (*entity.Response, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId, "index": index})

	run, err := m.run(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	row, found := run.Row(index)
	if !found || row.ResponseKey == "" {
		return nil, usecase.RejectUnknownRun(runId)
	}
	response, err := m.findResponse.Find(ctx, row.ResponseKey)
	if errs.IsNotFoundError(err) {
		return nil, usecase.RejectUnknownRun(runId)
	}
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "index": index, "status": response.StatusCode})
	return response, nil
}

func (m *manageRunner) Export(ctx context.Context, localId, runId string, format enums.RunExportFormat, path string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId, "format": format, "path": path})

	if _, active := m.manageActiveRuns.Find(ctx, localId, runId); active {
		return usecase.RejectActiveRun(runId)
	}
	data, err := m.exported(ctx, localId, runId, format)
	if err != nil {
		return err
	}
	if err := m.saveRunExport.SaveRunExport(ctx, path, data); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "format": format, "path": path, "bytes": len(data)})
	return nil
}

func (m *manageRunner) exported(ctx context.Context, localId, runId string, format enums.RunExportFormat) ([]byte, error) {
	if format != runexportformat.Junit {
		return m.findRuns.FindRunFile(ctx, localId, runId)
	}
	run, err := m.findRuns.FindRun(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	if run.Config.IsPerformance() {
		return nil, usecase.RejectInvalidRunSetting("format", "JUnit is not available for performance runs")
	}
	return usecase.BuildRunJUnit(*run)
}

func (m *manageRunner) run(ctx context.Context, localId, runId string) (*entity.Run, error) {
	if active, found := m.manageActiveRuns.Find(ctx, localId, runId); found {
		return &active.Run, nil
	}
	return m.findRuns.FindRun(ctx, localId, runId)
}

func (m *manageRunner) releasedResponses(ctx context.Context, run entity.Run) ([]string, error) {
	released := make([]string, 0)
	for _, key := range run.ResponseKeys() {
		_, err := m.findResponse.Find(ctx, key)
		if errs.IsNotFoundError(err) {
			released = append(released, key)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return released, nil
}
