package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

type exportRun struct {
	findActiveRuns usecase.FindActiveRuns
	findRuns       usecase.FindRuns
	saveRunExport  usecase.SaveRunExport
}

func NewExportRunService(findActiveRuns usecase.FindActiveRuns, findRuns usecase.FindRuns, saveRunExport usecase.SaveRunExport) adapter.ExportRunService {
	return &exportRun{findActiveRuns: findActiveRuns, findRuns: findRuns, saveRunExport: saveRunExport}
}

func (e *exportRun) Export(ctx context.Context, localId, runId string, format enums.RunExportFormat, path string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId, "format": format, "path": path})

	if active, found := e.findActiveRuns.FindById(ctx, localId, runId); found {
		return active.ActiveConflict()
	}
	run, err := e.findRuns.FindRun(ctx, localId, runId)
	if err != nil {
		return err
	}
	if err := run.ValidateExportAs(format); err != nil {
		return err
	}
	if err := e.saveRunExport.SaveRunExport(ctx, path, *run, format); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "format": format, "path": path})
	return nil
}
