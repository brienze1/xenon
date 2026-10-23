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

func (e *runExecutor) runRow(ctx context.Context, run *functionalRun, index, iteration int, row entity.RunPlanRow) ranRow {
	result := entity.NewRunRowResult(index, iteration, row)
	stored, err := e.findRequest.Find(ctx, run.root, row.RequestID)
	if errs.IsNotFoundError(err) {
		return ranRow{result: result.Failed(rowError(row.MissingRequest()))}
	}
	if err != nil {
		return ranRow{result: result.Failed(rowError(err))}
	}
	alias, _ := stored.Alias(row.AliasID)
	request := stored.Version(row.AliasID)
	result = result.ForRequest(request, alias)
	console := &rowConsole{entries: []entity.ConsoleEntry{}}
	runContext := run.context(iteration, index)
	sent, err := e.pipeline.send(ctx, sendCall{localId: run.target.LocalID, aliasId: row.AliasID, request: request, appVersion: run.appVersion, acknowledgement: run.acknowledgement, collect: console.collect}, runContext)
	result = result.WithConsole(console.collected())
	if err != nil && ctx.Err() != nil {
		return ranRow{result: result.Cancelled()}
	}
	if err != nil {
		return ranRow{result: result.Failed(rowError(err)).WithSecretsMasked(run.mask())}
	}
	run.secrets = entity.MergeSecrets(run.secrets, sent.secrets)
	result = result.WithExecution(*sent.execution)
	ran := ranRow{next: sent.execution.NextRequest(), sendId: sent.execution.ID}
	if sent.execution.HasResponse() {
		result = result.WithResponseKey(runContext.ResponseKey)
		ran.latency = sent.execution.Latency()
	}
	ran.result = result.WithSecretsMasked(run.mask())
	return ran
}

func (r *functionalRun) context(iteration, index int) entity.RunContext {
	if r.config.IsPerformance() {
		return r.performanceContext(iteration)
	}
	return r.baseContext(iteration).ForFunctionalRow(r.config.Functional, r.data, index)
}

func (r *functionalRun) performanceContext(iteration int) entity.RunContext {
	return r.baseContext(iteration).ForPerformance(r.data)
}

func (r *functionalRun) baseContext(iteration int) entity.RunContext {
	return entity.NewRunContext(r.id, r.environment.id, r.jarId, r.variables, r.acknowledgement, iteration)
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
