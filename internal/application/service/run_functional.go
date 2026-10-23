package service

import (
	"context"
	"slices"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/idutils"
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
	data            entity.IterationData
	rows            []entity.RunPlanRow
	secrets         []string
	config          entity.RunConfig
	executed        int
}

func (m *manageRunner) runFunctional(ctx context.Context, run functionalRun) {
	defer run.control.Finished()
	logger.Info(ctx, "Started", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id})

	status := m.runIterations(ctx, &run)
	m.finish(context.WithoutCancel(ctx), run, status)

	logger.Info(ctx, "Finished", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id, "status": status})
}

func (m *manageRunner) runIterations(ctx context.Context, run *functionalRun) enums.RunStatus {
	for iteration := range run.config.Functional.Iterations {
		if status, ended := m.runIteration(ctx, run, iteration); ended {
			return status
		}
	}
	return runstatus.Completed
}

func (m *manageRunner) runIteration(ctx context.Context, run *functionalRun, iteration int) (enums.RunStatus, bool) {
	position := 0
	for position < len(run.rows) {
		if !waitDelay(ctx, run.config.Functional.DelayBefore(run.executed)) {
			return runstatus.Stopped, true
		}
		if !waitWhilePaused(ctx, run.control) {
			return runstatus.Stopped, true
		}
		row := run.rows[position]
		m.change(ctx, *run, func(active *entity.Run) { active.Starting(row, iteration, timeutils.Now()) })
		ran := m.runRow(ctx, run, run.executed, iteration, row)
		run.executed++
		next, known := entity.NextRunPosition(run.rows, position, ran.next)
		if !known {
			ran.result = m.warnUnknownNextRequest(ctx, *run, ran)
		}
		position = next
		if recorded, found := m.change(ctx, *run, func(active *entity.Run) { active.Record(ran.result, timeutils.Now()) }); found {
			m.publishRunEvent.Progress(ctx, recorded.Run)
		}
		if ctx.Err() != nil {
			return runstatus.Stopped, true
		}
		if run.config.Functional.StopsAfter(ran.result) {
			return runstatus.StoppedOnError, true
		}
	}
	return "", false
}

func (m *manageRunner) warnUnknownNextRequest(ctx context.Context, run functionalRun, ran ranRow) entity.RunRowResult {
	warning := entity.NewConsoleWarningEntry(idutils.NewUlid(), ran.sendId, entity.UnknownNextRequestWarning(ran.next.Target), timeutils.Now())
	if run.config.Functional.DisableLogs {
		return ran.result
	}
	m.core.console.append(ctx, warning)
	return ran.result.WithConsoleEntry(warning)
}

func (m *manageRunner) finish(ctx context.Context, run functionalRun, status enums.RunStatus) {
	defer m.manageCookieJar.DropJar(ctx, run.jarId)
	if run.control.IsAbandoned() {
		m.abandon(ctx, run)
		return
	}
	if err := m.keepCookies(ctx, run); err != nil {
		logger.Warn(ctx, err, "The cookies of the run could not be saved", map[string]any{"runId": run.id})
	}
	mask := run.mask()
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
	pruned, err := m.saveRun.SaveRun(ctx, finished.Run)
	if err != nil {
		logger.Warn(ctx, err, "The finished run could not be stored", map[string]any{"runId": run.id})
	}
	m.releaseResponses(ctx, pruned...)
	m.manageActiveRuns.Release(ctx, run.target.LocalID, run.id)
	m.publishRunEvent.Finished(ctx, finished.Run)
}

func (m *manageRunner) abandon(ctx context.Context, run functionalRun) {
	if active, found := m.manageActiveRuns.Find(ctx, run.target.LocalID, run.id); found {
		m.releaseResponses(ctx, active.Run)
	}
	m.manageActiveRuns.Release(ctx, run.target.LocalID, run.id)
}

func (m *manageRunner) keepCookies(ctx context.Context, run functionalRun) error {
	if !run.config.Functional.SaveCookies {
		return nil
	}
	localId := run.target.LocalID
	if err := m.core.cookies.load(ctx, localId); err != nil {
		return err
	}
	before := m.core.cookies.revision(ctx, localId)
	merged := entity.MergeCookies(m.manageCookieJar.JarState(ctx, localId).Cookies, m.manageCookieJar.JarState(ctx, run.jarId).Cookies)
	m.manageCookieJar.StoreCookies(ctx, localId, merged)
	return m.core.cookies.settle(ctx, localId, before)
}

func (m *manageRunner) releaseResponses(ctx context.Context, runs ...entity.Run) {
	for _, run := range runs {
		for _, key := range run.ResponseKeys() {
			if err := m.releaseResponse.Release(ctx, key); err != nil {
				logger.Warn(ctx, err, "A kept response of a run could not be released", map[string]any{"runId": run.ID, "key": key})
			}
		}
	}
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

func waitDelay(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
