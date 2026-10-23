package service

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type rowConsole struct {
	entries []entity.ConsoleEntry
	mutex   sync.Mutex
}

func (m *manageRunner) runRow(ctx context.Context, run *functionalRun, index, iteration int, row entity.RunPlanRow) entity.RunRowResult {
	result := entity.NewRunRowResult(index, iteration, row)
	stored, err := m.findRequest.Find(ctx, run.root, row.RequestID)
	if err != nil {
		return result.Failed(rowError(usecase.NewMissingRunRequest(row.RequestID)))
	}
	alias, _ := stored.Alias(row.AliasID)
	request := stored.Version(row.AliasID)
	result = result.ForRequest(request, alias)
	console := &rowConsole{entries: []entity.ConsoleEntry{}}
	sent, err := m.core.send(ctx, sendCall{localId: run.target.LocalID, aliasId: row.AliasID, request: request, appVersion: run.appVersion, acknowledgement: run.acknowledgement}, run.context(iteration, console))
	result = result.WithConsole(console.collected())
	if err != nil && ctx.Err() != nil {
		return result.Cancelled()
	}
	if err != nil {
		return result.Failed(rowError(err)).WithSecretsMasked(run.mask())
	}
	run.remember(sent.secrets)
	return result.WithExecution(*sent.execution).WithSecretsMasked(run.mask())
}

func (r *functionalRun) context(iteration int, console *rowConsole) entity.RunContext {
	runContext := entity.RunContext{
		RunID:           r.id,
		EnvironmentID:   r.environment.id,
		CookieJarID:     r.jarId,
		Variables:       r.variables,
		Acknowledgement: r.acknowledgement,
		Iteration:       iteration,
		IterationCount:  r.config.Functional.Iterations,
		MuteConsole:     r.config.Functional.DisableLogs,
	}
	if !r.config.Functional.DisableLogs {
		runContext.Collect = console.collect
	}
	return runContext
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
