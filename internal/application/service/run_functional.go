package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type functionalRun struct {
	control         *runControl
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

func (e *runExecutor) runFunctional(ctx context.Context, run functionalRun) {
	defer run.control.finished()
	logger.Info(ctx, "Started", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id})

	status := e.runIterations(ctx, &run)
	e.finish(context.WithoutCancel(ctx), run, status)

	logger.Info(ctx, "Finished", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id, "status": status})
}

func (e *runExecutor) runIterations(ctx context.Context, run *functionalRun) enums.RunStatus {
	for iteration := range run.config.Functional.Iterations {
		if status, ended := e.runIteration(ctx, run, iteration); ended {
			return status
		}
	}
	return runstatus.Completed
}

func (e *runExecutor) runIteration(ctx context.Context, run *functionalRun, iteration int) (enums.RunStatus, bool) {
	position := 0
	for position < len(run.rows) {
		if !waitDelay(ctx, run.config.Functional.DelayBefore(run.executed)) {
			return runstatus.Stopped, true
		}
		if !waitWhilePaused(ctx, run.control) {
			return runstatus.Stopped, true
		}
		row := run.rows[position]
		e.change(ctx, *run, func(active *entity.Run) { active.Starting(row, iteration, timeutils.Now()) })
		ran := e.runRow(ctx, run, run.executed, iteration, row)
		run.executed++
		next, known := entity.NextRunPosition(run.rows, position, ran.next)
		if !known {
			ran.result = e.warnUnknownNextRequest(ctx, *run, ran)
		}
		position = next
		if recorded, found := e.change(ctx, *run, func(active *entity.Run) { active.Record(ran.result, timeutils.Now()) }); found {
			if err := e.publishRunEvent.Progress(ctx, *recorded); err != nil {
				logger.Warn(ctx, err, "The run progress could not be published", map[string]any{"runId": run.id})
			}
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

func (e *runExecutor) warnUnknownNextRequest(ctx context.Context, run functionalRun, ran ranRow) entity.RunRowResult {
	warning := entity.NewConsoleWarningEntry(idutils.NewUlid(), ran.sendId, entity.UnknownNextRequestWarning(ran.next.Target), timeutils.Now())
	if !run.config.LogsToConsole() {
		return ran.result
	}
	e.pipeline.console.append(ctx, warning)
	return ran.result.WithConsoleEntry(warning)
}

func (e *runExecutor) finish(ctx context.Context, run functionalRun, status enums.RunStatus) {
	defer e.manageCookieJar.DropJar(ctx, run.jarId)
	if run.control.isAbandoned() {
		e.abandon(ctx, run)
		return
	}
	cookiesErr := e.keepCookies(ctx, run)
	if cookiesErr != nil {
		logger.Warn(ctx, cookiesErr, "The cookies of the run could not be saved", map[string]any{"runId": run.id})
	}
	mask := run.mask()
	finished, found := e.change(ctx, run, func(active *entity.Run) {
		if cookiesErr != nil {
			active.Notice(entity.UnsavedRunCookiesNotice)
		}
		active.MaskSecrets(mask)
		active.Finish(status, timeutils.Now())
	})
	if !found {
		return
	}
	pruned, err := e.store(ctx, *finished)
	if err != nil {
		logger.Warn(ctx, err, "The finished run could not be stored", map[string]any{"runId": run.id})
	}
	if err = releaseRunResponses(ctx, e.releaseResponse, pruned...); err != nil {
		logger.Warn(ctx, err, "The responses of the pruned runs could not be released", map[string]any{"runId": run.id})
	}
	e.releaseActiveRun.Release(ctx, run.target.LocalID, run.id)
	if err = e.publishRunEvent.Finished(ctx, *finished); err != nil {
		logger.Warn(ctx, err, "The finished run could not be published", map[string]any{"runId": run.id})
	}
}

func (e *runExecutor) store(ctx context.Context, run entity.Run) ([]entity.Run, error) {
	if err := e.saveRun.SaveRun(ctx, run); err != nil {
		return nil, err
	}
	stored, err := e.findRuns.FindRuns(ctx, run.Target.LocalID, run.Target.ItemID)
	if err != nil {
		return nil, err
	}
	pruned := make([]entity.Run, 0)
	for _, overflow := range entity.RunsBeyondStoredLimit(stored) {
		deleted, err := e.deleteRuns.DeleteRun(ctx, run.Target.LocalID, overflow.ID)
		if err != nil {
			return pruned, err
		}
		pruned = append(pruned, *deleted)
	}
	return pruned, nil
}

func (e *runExecutor) abandon(ctx context.Context, run functionalRun) {
	if active, found := e.findActiveRuns.FindById(ctx, run.target.LocalID, run.id); found {
		if err := releaseRunResponses(ctx, e.releaseResponse, *active); err != nil {
			logger.Warn(ctx, err, "The responses of the abandoned run could not be released", map[string]any{"runId": run.id})
		}
	}
	e.releaseActiveRun.Release(ctx, run.target.LocalID, run.id)
}

func (e *runExecutor) keepCookies(ctx context.Context, run functionalRun) error {
	if !run.config.SavesCookies() {
		return nil
	}
	localId := run.target.LocalID
	if err := e.pipeline.cookies.load(ctx, localId); err != nil {
		return err
	}
	before := e.pipeline.cookies.revision(ctx, localId)
	merged := entity.MergeCookies(e.manageCookieJar.JarState(ctx, localId).Cookies, e.manageCookieJar.JarState(ctx, run.jarId).Cookies)
	e.manageCookieJar.StoreCookies(ctx, localId, merged)
	return e.pipeline.cookies.settle(ctx, localId, before)
}

func (e *runExecutor) change(ctx context.Context, run functionalRun, change func(*entity.Run)) (*entity.Run, bool) {
	return e.changeActiveRun.Change(ctx, run.target.LocalID, run.id, change)
}

func waitWhilePaused(ctx context.Context, control *runControl) bool {
	for {
		resumed, paused := control.awaitResume()
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
