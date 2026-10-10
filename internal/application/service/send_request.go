package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type sendRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	saveHistoryEntry      usecase.SaveHistoryEntry
	generateDynamicValue  usecase.GenerateDynamicValue
	variables             requestVariables
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	generateDynamicValue usecase.GenerateDynamicValue,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
) adapter.SendRequestService {
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		sendHttpRequest:       sendHttpRequest,
		saveResponse:          saveResponse,
		saveHistoryEntry:      saveHistoryEntry,
		generateDynamicValue:  generateDynamicValue,
		variables: requestVariables{
			findTreeItem:       findTreeItem,
			loadWorkspace:      loadWorkspace,
			findEnvironment:    findEnvironment,
			findLocalVariables: findLocalVariables,
			findVault:          findVault,
		},
	}
}

func (s *sendRequest) Send(ctx context.Context, localId, tabId string, request entity.Request, appVersion string, acknowledgement entity.SendAcknowledgement) (*entity.Response, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "tabId": tabId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.load(ctx, localId, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	preflight := usecase.CheckSendPreflight(request, scopes.ancestors, scopes.context, scopes.environment)
	if preflight.AwaitsConfirmation(acknowledgement) {
		return nil, errs.ConflictError("the active environment asks for a confirmation before every send", "SND-05409", usecase.ConfirmationDetails(preflight)...)
	}
	if preflight.AwaitsSecretsAcknowledgement(acknowledgement) {
		return nil, errs.ConflictError("the request uses secrets that have no value on this machine or are locked", "SND-06409", usecase.UnavailableSecretDetails(preflight)...)
	}
	secrets := scopes.context.SecretValues()
	ctx = logger.WithRedactions(ctx, secrets...)
	sentAt := timeutils.Now()
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(request, scopes.ancestors, scopes.context, func(name string) (string, bool) {
		return s.generateDynamicValue.Generate(ctx, name)
	})
	effective := usecase.ResolveEffectiveRequest(resolved, resolvedAncestors, workspace.Path)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, appVersion))
	response := s.sendHttpRequest.Send(ctx, effective).WithMaskedValues(secrets)
	if err := s.saveResponse.Save(ctx, tabId, response); err != nil {
		return nil, err
	}
	if err := s.saveHistoryEntry.Save(ctx, localId, entity.NewHistoryEntry(idutils.NewUlid(), sentAt, request, response)); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "tabId": tabId, "request": request, "appVersion": appVersion, "result": response})
	return &response, nil
}
