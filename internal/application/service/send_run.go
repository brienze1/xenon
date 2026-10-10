package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	maxRunRequestDepth = 10
	runRequestNotFound = "runRequest: no saved request with id "
	runRequestTooDeep  = "runRequest: requests run from scripts are nested too deep"
)

// sendRun is one send with everything its scripts reach: the scope values, the settings, and the requests they send or run.
type sendRun struct {
	service    *sendRequest
	scopes     *requestScopes
	localId    string
	root       string
	appVersion string
	secrets    []string
	settings   entity.AppSettings
	depth      int
}

func (r sendRun) execute(ctx context.Context, tabId string, request entity.Request, ancestors []entity.Collection, owners []entity.ScriptOwner, execution *entity.SendExecution) error {
	location := requestScopes{ancestors: ancestors}.location(request)
	r.runScripts(ctx, execution, request, location, usecase.CollectScriptChain(owners, scriptphase.PreRequest))
	if execution.ShouldSend() {
		if err := r.send(ctx, tabId, request, ancestors, execution); err != nil {
			return err
		}
	}
	if execution.RunsPostScripts() {
		r.runScripts(ctx, execution, request, location, usecase.CollectScriptChain(owners, scriptphase.PostResponse))
	}
	return nil
}

func (r sendRun) runScripts(ctx context.Context, execution *entity.SendExecution, request entity.Request, location []string, chain []entity.ScriptRef) {
	for _, script := range chain {
		before, vaultBefore := execution.Scopes, cloneVault(execution.Vault)
		cookies := r.cookies(execution)
		base := entity.ScriptContext{Info: entity.NewScriptInfo(request.ID, request.Name, script.OwnerName, location), Cookies: cookies, ResponseCookies: cookies, Dynamic: r.dynamic(ctx), Console: r.console(ctx, execution.ID, script)}
		run := r.service.runScript.Run(ctx, script, execution.ScriptContext(base), r, r.settings)
		execution.AddRun(run)
		r.service.appendConsoleEntries.Append(ctx, r.consoleEntries(execution.ID, run))
		if err := r.service.scriptVariables.persist(ctx, r.localId, r.root, r.scopes.refs(), before, run.Scopes, vaultBefore, run.Vault); err != nil {
			logger.Warn(ctx, err, "The variable changes of a script could not be saved", map[string]any{"script": script.Label()})
			r.service.appendConsoleEntries.Append(ctx, []entity.ConsoleEntry{entity.NewConsoleErrorEntry(idutils.NewUlid(), execution.ID, entity.NewUnsavedVariablesError(script, err.Error()), timeutils.Now())})
		}
		if run.StopsSend() {
			return
		}
	}
}

func (r sendRun) send(ctx context.Context, tabId string, request entity.Request, ancestors []entity.Collection, execution *entity.SendExecution) error {
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(execution.Request, ancestors, r.variables(execution), r.dynamic(ctx))
	effective := usecase.ResolveEffectiveRequest(resolved, resolvedAncestors, r.root)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, r.appVersion))
	response := r.service.sendHttpRequest.Send(ctx, effective).WithMaskedValues(r.secrets).WithSentRequest(request)
	if tabId != "" {
		if err := r.service.saveResponse.Save(ctx, tabId, response); err != nil {
			return err
		}
	}
	execution.RecordResponse(effective, response)
	return nil
}

func (r sendRun) variables(execution *entity.SendExecution) entity.VariableContext {
	return execution.Variables(r.scopes.context).WithVault(execution.Vault)
}

func (r sendRun) cookies(execution *entity.SendExecution) []entity.ScriptCookie {
	if execution.Response == nil || execution.SentRequest == nil {
		return []entity.ScriptCookie{}
	}
	return entity.NewScriptCookies(execution.Response.Headers, execution.SentRequest.Host())
}

func (r sendRun) dynamic(ctx context.Context) entity.DynamicValue {
	return func(name string) (string, bool) {
		return r.service.generateDynamicValue.Generate(ctx, name)
	}
}

func (r sendRun) console(ctx context.Context, sendId string, script entity.ScriptRef) entity.ScriptLogSink {
	return func(log entity.ScriptLog) {
		if log.IsClear() {
			return
		}
		entry := entity.NewConsoleLogEntry(idutils.NewUlid(), sendId, log.Level, script, log.Text(), log.Arguments, timeutils.Now()).WithSecretsMasked(r.mask)
		r.service.appendConsoleEntries.Append(ctx, []entity.ConsoleEntry{entry})
	}
}

func (r sendRun) consoleEntries(sendId string, run entity.ScriptRun) []entity.ConsoleEntry {
	if !run.HasFailed() {
		return []entity.ConsoleEntry{}
	}
	return []entity.ConsoleEntry{entity.NewConsoleErrorEntry(idutils.NewUlid(), sendId, *run.Error, timeutils.Now()).WithSecretsMasked(r.mask)}
}

func (r sendRun) mask(text string) (string, bool) {
	return usecase.MaskSecrets(text, r.secrets)
}

// Send is pm.sendRequest: the request resolves with the scopes as the script left them, shares the main send's
// confirmation, rejects secrets without a value and never reaches History.
func (r sendRun) Send(ctx context.Context, request entity.Request, scopes entity.ScriptScopes) entity.NestedRequest {
	logger.Info(ctx, "Started", map[string]any{"request": request})

	variables := r.scopes.context.WithScriptScopes(scopes)
	preflight := usecase.CheckSendPreflight(request, variables, nil)
	if preflight.HasUnavailableSecrets() {
		rejected := entity.NewRejectedNestedRequest(preflight.UnavailableSecretsMessage())
		logger.Info(ctx, "Finished", map[string]any{"request": request, "result": rejected})
		return rejected
	}
	resolved, _ := usecase.ResolveRequestVariables(request, nil, variables, r.dynamic(ctx))
	effective := usecase.ResolveEffectiveRequest(resolved, nil, r.root)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, r.appVersion))
	sent := entity.NewSentNestedRequest(effective, r.service.sendHttpRequest.Send(ctx, effective).WithMaskedValues(r.secrets))

	logger.Info(ctx, "Finished", map[string]any{"request": request, "result": sent})
	return sent
}

// Run is pm.execution.runRequest: the saved request runs with its own scripts on the caller's scopes and hands back
// the response and the scopes as its scripts left them; it never reaches History.
func (r sendRun) Run(ctx context.Context, requestId string, scopes entity.ScriptScopes) entity.NestedRequest {
	logger.Info(ctx, "Started", map[string]any{"requestId": requestId})

	if r.depth >= maxRunRequestDepth {
		return entity.NewMissingRunNestedRequest(requestId, runRequestTooDeep)
	}
	request, ancestors, err := r.savedRequest(ctx, requestId)
	if err != nil {
		logger.Info(ctx, "Finished", map[string]any{"requestId": requestId, "error": err.Error()})
		return entity.NewMissingRunNestedRequest(requestId, runRequestNotFound+requestId)
	}
	child := r
	child.depth++
	execution := entity.NewSendExecution(idutils.NewUlid())
	execution.Start(*request, scopes, cloneVault(r.scopes.vault))
	if err := child.execute(ctx, "", *request, ancestors, request.ScriptOwners(ancestors), &execution); err != nil {
		return entity.NewMissingRunNestedRequest(requestId, err.Error())
	}
	nested := entity.NewRunNestedRequest(requestId, execution)

	logger.Info(ctx, "Finished", map[string]any{"requestId": requestId, "result": nested})
	return nested
}

func (r sendRun) savedRequest(ctx context.Context, requestId string) (*entity.Request, []entity.Collection, error) {
	if !entity.IsSavedRequestID(requestId) {
		return nil, nil, errs.NotFoundError("not a saved request id", "SND-15404")
	}
	request, err := r.service.findRequest.Find(ctx, r.root, requestId)
	if err != nil {
		return nil, nil, err
	}
	ancestors, err := r.service.variables.findTreeItem.FindAncestors(ctx, r.root, requestId)
	if err != nil {
		return nil, nil, err
	}
	return request, ancestors, nil
}
