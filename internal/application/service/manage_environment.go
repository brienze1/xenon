package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageEnvironment struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	listEnvironments      usecase.ListEnvironments
	findEnvironment       usecase.FindEnvironment
	findLocalVariables    usecase.FindLocalVariables
	saveLocalVariables    usecase.SaveLocalVariables
}

func NewManageEnvironmentService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	listEnvironments usecase.ListEnvironments,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
) adapter.ManageEnvironmentService {
	return &manageEnvironment{
		findWorkspaceRegistry: findWorkspaceRegistry,
		listEnvironments:      listEnvironments,
		findEnvironment:       findEnvironment,
		findLocalVariables:    findLocalVariables,
		saveLocalVariables:    saveLocalVariables,
	}
}

func (m *manageEnvironment) List(ctx context.Context, localId string) (*entity.EnvironmentList, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	environments, err := m.listEnvironments.List(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	local, err := m.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return nil, err
	}
	list := entity.NewEnvironmentList(environments, local.ActiveEnvironmentID)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": list})
	return &list, nil
}

func (m *manageEnvironment) SetActive(ctx context.Context, localId, id string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", err
	}
	local, err := m.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return "", err
	}
	updated := local.WithActiveEnvironment(id)
	if updated.HasActiveEnvironment() {
		if _, err := m.findEnvironment.Find(ctx, workspace.Path, id); err != nil {
			return "", err
		}
	}
	if err := m.saveLocalVariables.Save(ctx, localId, updated); err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": updated.ActiveEnvironmentID})
	return updated.ActiveEnvironmentID, nil
}
