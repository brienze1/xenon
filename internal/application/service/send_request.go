package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type sendRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	findAppSettings       usecase.FindAppSettings
	runScript             usecase.RunScript
	appendConsoleEntries  usecase.AppendConsoleEntries
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	saveHistoryEntry      usecase.SaveHistoryEntry
	generateDynamicValue  usecase.GenerateDynamicValue
	variables             requestVariables
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	findRequest usecase.FindRequest,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	findAppSettings usecase.FindAppSettings,
	runScript usecase.RunScript,
	appendConsoleEntries usecase.AppendConsoleEntries,
	generateDynamicValue usecase.GenerateDynamicValue,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
) adapter.SendRequestService {
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		findAppSettings:       findAppSettings,
		runScript:             runScript,
		appendConsoleEntries:  appendConsoleEntries,
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

func (s *sendRequest) Send(ctx context.Context, localId, tabId string, request entity.Request, appVersion string, acknowledgement entity.SendAcknowledgement) (*entity.SendExecution, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "tabId": tabId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.load(ctx, localId, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	inheriting := request.Inheriting(usecase.ResolveEffectiveRequest(request, scopes.ancestors, workspace.Path))
	preflight := usecase.CheckSendPreflight(inheriting, scopes.context, scopes.environment)
	if preflight.AwaitsConfirmation(acknowledgement) {
		return nil, usecase.RejectUnconfirmedSend(preflight)
	}
	if preflight.AwaitsSecretsAcknowledgement(acknowledgement) {
		return nil, usecase.RejectUnavailableSecrets(preflight)
	}
	owners, err := s.scriptOwners(ctx, workspace.Path, request, scopes.ancestors)
	if err != nil {
		return nil, err
	}
	settings, err := s.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	secrets := scopes.context.SecretValues()
	ctx = logger.WithRedactions(ctx, secrets...)
	sentAt := timeutils.Now()
	execution := entity.NewSendExecution(idutils.NewUlid())
	execution.Start(request)
	pipeline := scriptPipeline{
		runScript:            s.runScript,
		appendConsoleEntries: s.appendConsoleEntries,
		requests: scriptRequests{
			sendHttpRequest:      s.sendHttpRequest,
			generateDynamicValue: s.generateDynamicValue,
			variables:            scopes.context,
			root:                 workspace.Path,
			appVersion:           appVersion,
			secrets:              secrets,
		},
		variables: scopes.context,
		settings:  *settings,
		secrets:   secrets,
	}
	pipeline.run(ctx, &execution, usecase.CollectScriptChain(owners, scriptphase.PreRequest))
	if execution.ShouldSend() {
		if err := s.send(ctx, tabId, request, &execution, scopes, workspace.Path, appVersion, secrets); err != nil {
			return nil, err
		}
	}
	if execution.RunsPostScripts() {
		pipeline.run(ctx, &execution, usecase.CollectScriptChain(owners, scriptphase.PostResponse))
	}
	if err := s.saveHistoryEntry.Save(ctx, localId, entity.NewHistoryEntry(idutils.NewUlid(), sentAt, request, execution)); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "tabId": tabId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement, "result": execution})
	return &execution, nil
}

func (s *sendRequest) scriptOwners(ctx context.Context, root string, request entity.Request, ancestors []entity.Collection) ([]entity.ScriptOwner, error) {
	if !entity.IsSavedRequestID(request.ID) {
		return request.ScriptOwners(ancestors), nil
	}
	stored, err := s.findRequest.Find(ctx, root, request.ID)
	if err != nil {
		return nil, err
	}
	return request.WithScriptsOf(*stored).ScriptOwners(ancestors), nil
}

func (s *sendRequest) send(ctx context.Context, tabId string, request entity.Request, execution *entity.SendExecution, scopes *requestScopes, root, appVersion string, secrets []string) error {
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(execution.Request, scopes.ancestors, execution.Variables(scopes.context), func(name string) (string, bool) {
		return s.generateDynamicValue.Generate(ctx, name)
	})
	effective := usecase.ResolveEffectiveRequest(resolved, resolvedAncestors, root)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, appVersion))
	response := s.sendHttpRequest.Send(ctx, effective).WithMaskedValues(secrets).WithSentRequest(request)
	if err := s.saveResponse.Save(ctx, tabId, response); err != nil {
		return err
	}
	execution.RecordResponse(effective, response)
	return nil
}
