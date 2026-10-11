package service

import (
	"context"
	"slices"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	maxRunRequestDepth = 10
	runRequestNotFound = "runRequest: no saved request with id "
	runRequestTooDeep  = "runRequest: requests run from scripts are nested too deep"
)

type sendRun struct {
	service    *sendRequest
	scopes     *requestScopes
	caller     *entity.ScriptRef
	token      *entity.OAuth2Token
	localId    string
	tabId      string
	root       string
	appVersion string
	sendId     string
	secrets    []string
	settings   entity.AppSettings
	depth      int
}

func (r sendRun) execute(ctx context.Context, tabId string, request entity.Request, ancestors []entity.Collection, owners []entity.ScriptOwner, execution *entity.SendExecution) error {
	location := request.Location(ancestors)
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
		cookies := execution.ScriptCookies()
		base := entity.ScriptContext{Info: entity.NewScriptInfo(request.ID, request.Name, script.OwnerName, location), Cookies: cookies, ResponseCookies: cookies, Dynamic: r.dynamic(ctx), Console: r.console(ctx, execution, script)}
		run := r.service.runScript.Run(ctx, script, execution.ScriptContext(base), r.calledBy(script), r, r.settings)
		mask := r.mask(r.secretsOf(execution))
		execution.AddRun(run)
		r.service.console.append(ctx, r.runEntries(run, mask)...)
		if err := r.service.scriptVariables.persist(ctx, r.localId, r.root, entity.ScriptScopeRefs(r.scopes.environment, r.scopes.ancestors), before, run.Scopes, vaultBefore, run.Vault); err != nil {
			logger.Warn(ctx, err, "The variable changes of a script could not be saved", map[string]any{"script": script.Label()})
			r.service.console.append(ctx, entity.NewConsoleErrorEntry(idutils.NewUlid(), r.sendId, entity.NewUnsavedVariablesError(script, err.Error()), timeutils.Now()).WithSecretsMasked(mask))
		}
		if run.StopsSend() {
			return
		}
	}
}

func (r sendRun) send(ctx context.Context, tabId string, request entity.Request, ancestors []entity.Collection, execution *entity.SendExecution) error {
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(execution.Request, ancestors, r.variables(execution), r.dynamic(ctx))
	effective := usecase.ResolveEffectiveRequest(resolved, resolvedAncestors, r.root).WithCookieJar(r.localId)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, r.appVersion, r.service.cookies.header(ctx, r.localId, effective))).WithAuthOwner(r.tabId)
	effective, err := r.authorize(ctx, effective)
	if err != nil {
		return err
	}
	effective, response, err := r.exchange(ctx, effective)
	if err != nil {
		return err
	}
	response = response.WithMaskedValues(r.secrets).WithSentRequest(request)
	if tabId != "" {
		if err := r.service.saveResponse.Save(ctx, tabId, response); err != nil {
			return err
		}
	}
	execution.RecordResponse(effective, response)
	r.service.console.append(ctx, r.requestEntry(effective, response, r.depth > 0).WithSecretsMasked(r.mask(r.secretsOf(execution))))
	return nil
}

func (r sendRun) authorize(ctx context.Context, request entity.EffectiveRequest) (entity.EffectiveRequest, error) {
	auth := request.Auth
	switch {
	case auth.IsApiKey():
		return usecase.ApplyApiKey(request, auth.ApiKey), nil
	case auth.IsOAuth2():
		return usecase.ApplyOAuthToken(request, auth.OAuth2, r.token), nil
	case auth.IsAwsV4():
		signed, err := r.service.signAwsRequest.Sign(ctx, request, auth.AwsV4)
		if err != nil {
			return request, err
		}
		return *signed, nil
	default:
		return request, nil
	}
}

func (r sendRun) exchange(ctx context.Context, request entity.EffectiveRequest) (entity.EffectiveRequest, entity.Response, error) {
	if !request.Auth.IsDigest() {
		return request, r.service.sendHttpRequest.Send(ctx, request), nil
	}
	digest := request.Auth.Digest
	state, err := r.service.findDigestState.FindDigestState(ctx, r.localId, request.AuthOwner)
	if err != nil {
		return request, entity.Response{}, err
	}
	if typed, found := digest.TypedState(); state == nil && found {
		state = &typed
	}
	if state != nil {
		if request, err = r.answerDigest(ctx, request, *state); err != nil {
			return request, entity.Response{}, err
		}
	}
	response := r.service.sendHttpRequest.Send(ctx, request)
	challenge, challenged := response.DigestChallenge()
	if !challenged || !digest.RetriesChallenge() {
		return request, response, nil
	}
	r.service.console.append(ctx, r.requestEntry(request, response.WithMaskedValues(r.secrets), false).WithSecretsMasked(r.mask(r.secrets)))
	fresh, err := r.service.answerDigest.ReadChallenge(ctx, challenge)
	if err != nil {
		return request, response, nil
	}
	if request, err = r.answerDigest(ctx, request, *fresh); err != nil {
		return request, entity.Response{}, err
	}
	return request, r.service.sendHttpRequest.Send(ctx, request), nil
}

func (r sendRun) answerDigest(ctx context.Context, request entity.EffectiveRequest, state entity.DigestState) (entity.EffectiveRequest, error) {
	next := state.Next()
	header, err := r.service.answerDigest.Answer(ctx, request, request.Auth.Digest, next)
	if err != nil {
		return request, err
	}
	if err := r.service.saveDigestState.SaveDigestState(ctx, r.localId, request.AuthOwner, next); err != nil {
		return request, err
	}
	return request.WithHeader(entity.HeaderAuthorization, header), nil
}

func (r sendRun) variables(execution *entity.SendExecution) entity.VariableContext {
	return execution.Variables(r.scopes.context).WithVault(execution.Vault)
}

func (r sendRun) dynamic(ctx context.Context) entity.DynamicValue {
	return func(name string) (string, bool) {
		return r.service.generateDynamicValue.Generate(ctx, name)
	}
}

func (r sendRun) console(ctx context.Context, execution *entity.SendExecution, script entity.ScriptRef) entity.ScriptLogSink {
	return func(log entity.ScriptLog) {
		if log.IsClear() {
			r.service.console.clear(ctx)
			return
		}
		secrets := r.secretsOf(execution)
		if log.HasScopes() {
			secrets = append(secrets, r.scopes.context.WithScriptScopes(*log.Scopes).SecretValues()...)
		}
		entry := entity.NewConsoleLogEntry(idutils.NewUlid(), r.sendId, script, log, timeutils.Now())
		r.service.console.append(ctx, entry.WithSecretsMasked(r.mask(secrets)))
	}
}

func (r sendRun) runEntries(run entity.ScriptRun, mask func(string) (string, bool)) []entity.ConsoleEntry {
	entries := make([]entity.ConsoleEntry, 0, len(run.Tests)+1)
	for _, test := range run.Tests {
		if test.IsFailed() {
			entries = append(entries, entity.NewConsoleTestFailureEntry(idutils.NewUlid(), r.sendId, test, timeutils.Now()).WithSecretsMasked(mask))
		}
	}
	if run.HasFailed() {
		entries = append(entries, entity.NewConsoleErrorEntry(idutils.NewUlid(), r.sendId, *run.Error, timeutils.Now()).WithSecretsMasked(mask))
	}
	return entries
}

func (r sendRun) requestEntry(request entity.EffectiveRequest, response entity.Response, nested bool) entity.ConsoleEntry {
	script := r.caller
	if !nested {
		script = nil
	}
	return entity.NewConsoleRequestEntry(idutils.NewUlid(), r.sendId, script, nested, request, response, timeutils.Now())
}

func (r sendRun) calledBy(script entity.ScriptRef) sendRun {
	r.caller = &script
	return r
}

func (r sendRun) secretsOf(execution *entity.SendExecution) []string {
	return append(slices.Clone(r.secrets), r.variables(execution).SecretValues()...)
}

func (r sendRun) mask(secrets []string) func(string) (string, bool) {
	return func(text string) (string, bool) {
		return usecase.MaskSecrets(text, secrets)
	}
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
	effective := usecase.ResolveEffectiveRequest(resolved, nil, r.root).WithCookieJar(r.localId)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, r.appVersion, r.service.cookies.header(ctx, r.localId, effective)))
	response := r.service.sendHttpRequest.Send(ctx, effective).WithMaskedValues(r.secrets)
	sent := entity.NewSentNestedRequest(effective, response)
	r.service.console.append(ctx, r.requestEntry(effective, response, true).WithSecretsMasked(r.mask(append(slices.Clone(r.secrets), variables.SecretValues()...))))

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
	if !entity.IsSavedRequestID(requestId) {
		return entity.NewMissingRunNestedRequest(requestId, runRequestNotFound+requestId)
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
