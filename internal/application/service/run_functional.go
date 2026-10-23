package service

import (
	"context"
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type functionalRun struct {
	control         *entity.RunControl
	variables       *entity.RunVariables
	target          entity.RunTarget
	environment     runEnvironment
	id              string
	root            string
	jarId           string
	appVersion      string
	acknowledgement entity.SendAcknowledgement
	rows            []entity.RunPlanRow
	secrets         []string
	config          entity.RunConfig
}

func (m *manageRunner) runFunctional(ctx context.Context, run functionalRun) {
	defer run.control.Finished()
	logger.Info(ctx, "Started", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id})

	status := m.runRows(ctx, &run)
	m.finish(context.WithoutCancel(ctx), run, status)

	logger.Info(ctx, "Finished", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id, "status": status})
}

func (m *manageRunner) runRows(ctx context.Context, run *functionalRun) enums.RunStatus {
	index := 0
	for iteration := range run.config.Functional.Iterations {
		for _, row := range run.rows {
			if !waitWhilePaused(ctx, run.control) {
				return runstatus.Stopped
			}
			m.change(ctx, *run, func(active *entity.Run) { active.Starting(row, iteration, timeutils.Now()) })
			result := m.runRow(ctx, run, index, iteration, row)
			index++
			if recorded, found := m.change(ctx, *run, func(active *entity.Run) { active.Record(result, timeutils.Now()) }); found {
				m.publishRunEvent.Progress(ctx, recorded.Run)
			}
			if ctx.Err() != nil {
				return runstatus.Stopped
			}
		}
	}
	return runstatus.Completed
}

func (m *manageRunner) finish(ctx context.Context, run functionalRun, status enums.RunStatus) {
	mask := run.mask()
	defer m.manageCookieJar.DropJar(ctx, run.jarId)
	finished, found := m.change(ctx, run, func(active *entity.Run) {
		rows := make([]entity.RunRowResult, 0, len(active.Rows))
		for _, row := range active.Rows {
			rows = append(rows, row.WithSecretsMasked(mask))
		}
		active.Rows = rows
		active.Finish(status, timeutils.Now())
	})
	if !found {
		return
	}
	if err := m.saveRun.SaveRun(ctx, finished.Run); err != nil {
		logger.Warn(ctx, err, "The finished run could not be stored", map[string]any{"runId": run.id})
	}
	m.manageActiveRuns.Release(ctx, run.target.LocalID, run.id)
	m.publishRunEvent.Finished(ctx, finished.Run)
}

func (m *manageRunner) change(ctx context.Context, run functionalRun, change func(*entity.Run)) (*entity.ActiveRun, bool) {
	return m.manageActiveRuns.Change(ctx, run.target.LocalID, run.id, change)
}

func (r *functionalRun) remember(secrets []string) {
	for _, secret := range secrets {
		if secret != "" && !slices.Contains(r.secrets, secret) {
			r.secrets = append(r.secrets, secret)
		}
	}
}

func waitWhilePaused(ctx context.Context, control *entity.RunControl) bool {
	for {
		resumed, paused := control.Resumed()
		if !paused {
			return ctx.Err() == nil
		}
		select {
		case <-resumed:
		case <-ctx.Done():
			return false
		}
	}
}
