package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
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
	scriptVariables       scriptVariables
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	saveTreeItem usecase.SaveTreeItem,
	findRequest usecase.FindRequest,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findEnvironment usecase.FindEnvironment,
	saveEnvironment usecase.SaveEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	findAppSettings usecase.FindAppSettings,
	runScript usecase.RunScript,
	appendConsoleEntries usecase.AppendConsoleEntries,
	generateDynamicValue usecase.GenerateDynamicValue,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
) adapter.SendRequestService {
	scopes := variableScopes{findEnvironment: findEnvironment, findTreeItem: findTreeItem, loadWorkspace: loadWorkspace}
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
		scriptVariables: scriptVariables{
			scopes:             scopes,
			rowStore:           variableRowStore{scopes: scopes, saveEnvironment: saveEnvironment, saveTreeItem: saveTreeItem, saveWorkspaceFile: saveWorkspaceFile},
			findLocalVariables: findLocalVariables,
			saveLocalVariables: saveLocalVariables,
			vaults: vaultWriter{
				findVault:            findVault,
				saveVault:            saveVault,
				findVaultSettings:    findVaultSettings,
				saveVaultSettings:    saveVaultSettings,
				probeKeychain:        probeKeychain,
				publishVariableEvent: publishVariableEvent,
			},
			publishVariableEvent: publishVariableEvent,
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
	execution.Start(request, scopes.scripts, scopes.vault)
	run := sendRun{service: s, scopes: scopes, localId: localId, root: workspace.Path, appVersion: appVersion, secrets: secrets, settings: *settings}
	if err := run.execute(ctx, tabId, request, scopes.ancestors, owners, &execution); err != nil {
		return nil, err
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
