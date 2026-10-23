package service

import (
	"context"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type rowConsole struct {
	entries []entity.ConsoleEntry
	mutex   sync.Mutex
}

type ranRow struct {
	next    *entity.NextRequest
	sendId  string
	result  entity.RunRowResult
	latency time.Duration
}

func (m *manageRunner) runRow(ctx context.Context, run *functionalRun, index, iteration int, row entity.RunPlanRow) ranRow {
	result := entity.NewRunRowResult(index, iteration, row)
	stored, err := m.findRequest.Find(ctx, run.root, row.RequestID)
	if err != nil {
		return ranRow{result: result.Failed(rowError(usecase.NewMissingRunRequest(row.RequestID)))}
	}
	alias, _ := stored.Alias(row.AliasID)
	request := stored.Version(row.AliasID)
	result = result.ForRequest(request, alias)
	console := &rowConsole{entries: []entity.ConsoleEntry{}}
	runContext := run.context(iteration, index, console)
	sent, err := m.core.send(ctx, sendCall{localId: run.target.LocalID, aliasId: row.AliasID, request: request, appVersion: run.appVersion, acknowledgement: run.acknowledgement}, runContext)
	result = result.WithConsole(console.collected())
	if err != nil && ctx.Err() != nil {
		return ranRow{result: result.Cancelled()}
	}
	if err != nil {
		return ranRow{result: result.Failed(rowError(err)).WithSecretsMasked(run.mask())}
	}
	run.remember(sent.secrets)
	result = result.WithExecution(*sent.execution)
	ran := ranRow{next: sent.execution.NextRequest(), sendId: sent.execution.ID}
	if sent.execution.Response != nil {
		result = result.WithResponseKey(runContext.ResponseKey)
		ran.latency = sent.execution.Response.Timings.Total
	}
	ran.result = result.WithSecretsMasked(run.mask())
	return ran
}

func (r *functionalRun) context(iteration, index int, console *rowConsole) entity.RunContext {
	if r.config.IsPerformance() {
		return r.performanceContext(iteration)
	}
	runContext := entity.RunContext{
		RunID:           r.id,
		EnvironmentID:   r.environment.id,
		CookieJarID:     r.jarId,
		Variables:       r.variables,
		Acknowledgement: r.acknowledgement,
		Iteration:       iteration,
		IterationCount:  r.config.Functional.Iterations,
		IterationData:   r.data.Row(iteration),
		MuteConsole:     r.config.Functional.DisableLogs,
	}
	if r.config.Functional.PersistResponses {
		runContext.ResponseKey = entity.RunResponseKey(r.id, index)
	}
	if !r.config.Functional.DisableLogs {
		runContext.Collect = console.collect
	}
	return runContext
}

func (r *functionalRun) performanceContext(iteration int) entity.RunContext {
	return entity.RunContext{
		RunID:           r.id,
		EnvironmentID:   r.environment.id,
		CookieJarID:     r.jarId,
		Variables:       r.variables,
		Acknowledgement: r.acknowledgement,
		Iteration:       iteration,
		IterationCount:  max(1, r.data.Count()),
		IterationData:   r.data.WrappedRow(iteration),
		MuteConsole:     true,
		Performance:     true,
	}
}

func (r *functionalRun) mask() func(string) (string, bool) {
	secrets := r.secrets
	return func(text string) (string, bool) { return usecase.MaskSecrets(text, secrets) }
}

func (c *rowConsole) collect(entry entity.ConsoleEntry) {
	c.mutex.Lock()
	c.entries = append(c.entries, entry.WithoutResponseBody())
	c.mutex.Unlock()
}

func (c *rowConsole) collected() []entity.ConsoleEntry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return append([]entity.ConsoleEntry{}, c.entries...)
}

func rowError(err error) string {
	if failure, isBase := err.(errs.BaseError); isBase {
		return failure.Description()
	}
	return err.Error()
}
