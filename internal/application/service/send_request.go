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
	findTreeItem          usecase.FindTreeItem
	loadWorkspace         usecase.LoadWorkspace
	findEnvironment       usecase.FindEnvironment
	findLocalVariables    usecase.FindLocalVariables
	findVault             usecase.FindVault
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	saveHistoryEntry      usecase.SaveHistoryEntry
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
) adapter.SendRequestService {
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findTreeItem:          findTreeItem,
		loadWorkspace:         loadWorkspace,
		findEnvironment:       findEnvironment,
		findLocalVariables:    findLocalVariables,
		findVault:             findVault,
		sendHttpRequest:       sendHttpRequest,
		saveResponse:          saveResponse,
		saveHistoryEntry:      saveHistoryEntry,
	}
}

func (s *sendRequest) Send(ctx context.Context, localId, tabId string, request entity.Request, appVersion string, acknowledgeMissingSecrets bool) (*entity.Response, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "tabId": tabId, "request": request, "appVersion": appVersion, "acknowledgeMissingSecrets": acknowledgeMissingSecrets})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	ancestors := make([]entity.Collection, 0)
	if request.HasID() {
		ancestors, err = s.findTreeItem.FindAncestors(ctx, workspace.Path, request.ID)
		if err != nil {
			return nil, err
		}
	}
	variables, err := s.variableContext(ctx, localId, workspace.Path, ancestors)
	if err != nil {
		return nil, err
	}
	preflight := usecase.CheckSendPreflight(request, ancestors, *variables)
	if preflight.HasUnavailableSecrets() && !acknowledgeMissingSecrets {
		return nil, errs.ConflictError("the request uses secrets that have no value on this machine or are locked", "SND-06409", usecase.UnavailableSecretDetails(preflight)...)
	}
	secrets := variables.SecretValues()
	ctx = logger.WithRedactions(ctx, secrets...)
	sentAt := timeutils.Now()
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(request, ancestors, *variables, nil)
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

func (s *sendRequest) variableContext(ctx context.Context, localId, root string, ancestors []entity.Collection) (*entity.VariableContext, error) {
	local, err := s.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return nil, err
	}
	environment, err := s.activeEnvironment(ctx, root, *local)
	if err != nil {
		return nil, err
	}
	workspace, err := s.loadWorkspace.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	vault, err := s.vault(ctx, localId, usecase.UsesSecretRows(environment, ancestors, workspace.Variables))
	if err != nil {
		return nil, err
	}
	variables := usecase.BuildVariableContext(environment, ancestors, workspace.Variables, *local, *vault)
	return &variables, nil
}

func (s *sendRequest) vault(ctx context.Context, localId string, needed bool) (*entity.Vault, error) {
	if !needed {
		return &entity.Vault{LocalID: localId}, nil
	}
	return s.findVault.Find(ctx, localId)
}

func (s *sendRequest) activeEnvironment(ctx context.Context, root string, local entity.LocalVariables) (*entity.Environment, error) {
	if !local.HasActiveEnvironment() {
		return nil, nil
	}
	environment, err := s.findEnvironment.Find(ctx, root, local.ActiveEnvironmentID)
	if errs.IsNotFoundError(err) {
		logger.Info(ctx, "The active environment no longer exists, sending without one", local)
		return nil, nil
	}
	return environment, err
}
