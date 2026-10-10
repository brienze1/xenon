package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type scriptPipeline struct {
	runScript            usecase.RunScript
	appendConsoleEntries usecase.AppendConsoleEntries
	requests             scriptRequests
	variables            entity.VariableContext
	secrets              []string
	settings             entity.AppSettings
}

func (p scriptPipeline) run(ctx context.Context, execution *entity.SendExecution, chain []entity.ScriptRef) {
	for _, script := range chain {
		run := p.runScript.Run(ctx, script, execution.ScriptContext(p.variables), p.requests, p.settings)
		execution.AddRun(run)
		p.appendConsoleEntries.Append(ctx, p.consoleEntries(execution.ID, run))
		if run.StopsSend() {
			return
		}
	}
}

func (p scriptPipeline) consoleEntries(sendId string, run entity.ScriptRun) []entity.ConsoleEntry {
	entries := make([]entity.ConsoleEntry, 0, len(run.Logs)+1)
	for _, log := range run.Logs {
		entries = append(entries, entity.NewConsoleLogEntry(idutils.NewUlid(), sendId, log.Level, run.Script, log.Text(), log.Arguments, timeutils.Now()).WithSecretsMasked(p.mask))
	}
	if run.HasFailed() {
		entries = append(entries, entity.NewConsoleErrorEntry(idutils.NewUlid(), sendId, *run.Error, timeutils.Now()).WithSecretsMasked(p.mask))
	}
	return entries
}

func (p scriptPipeline) mask(text string) (string, bool) {
	return usecase.MaskSecrets(text, p.secrets)
}
