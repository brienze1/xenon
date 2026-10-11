package service

import (
	"context"
	"slices"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	maxRunRequestDepth = 10
	runRequestNotFound = "runRequest: no saved request with id "
	runRequestTooDeep  = "runRequest: requests run from scripts are nested too deep"
)

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

	nested := r.run(ctx, requestId, scopes)

	logger.Info(ctx, "Finished", map[string]any{"requestId": requestId, "result": nested})
	return nested
}

func (r sendRun) run(ctx context.Context, requestId string, scopes entity.ScriptScopes) entity.NestedRequest {
	if r.depth >= maxRunRequestDepth {
		return entity.NewMissingRunNestedRequest(requestId, runRequestTooDeep)
	}
	if !entity.IsSavedRequestID(requestId) {
		return entity.NewMissingRunNestedRequest(requestId, runRequestNotFound+requestId)
	}
	request, ancestors, err := r.savedRequest(ctx, requestId)
	if err != nil {
		logger.Warn(ctx, err, "The saved request of a runRequest could not be read", map[string]any{"requestId": requestId})
		return entity.NewMissingRunNestedRequest(requestId, runRequestNotFound+requestId)
	}
	child := r
	child.depth++
	execution := entity.NewSendExecution(idutils.NewUlid())
	execution.Start(*request, scopes, cloneVault(r.scopes.vault))
	if err := child.execute(ctx, "", *request, ancestors, request.ScriptOwners(ancestors), &execution); err != nil {
		return entity.NewMissingRunNestedRequest(requestId, err.Error())
	}
	return entity.NewRunNestedRequest(requestId, execution)
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
