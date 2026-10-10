package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type setActiveEnvironment struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findEnvironment       usecase.FindEnvironment
	findLocalVariables    usecase.FindLocalVariables
	saveLocalVariables    usecase.SaveLocalVariables
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewSetActiveEnvironmentService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.SetActiveEnvironmentService {
	return &setActiveEnvironment{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findEnvironment:       findEnvironment,
		findLocalVariables:    findLocalVariables,
		saveLocalVariables:    saveLocalVariables,
		publishVariableEvent:  publishVariableEvent,
	}
}

func (s *setActiveEnvironment) SetActive(ctx context.Context, localId, id string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", err
	}
	local, err := s.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return "", err
	}
	updated := local.WithActiveEnvironment(id)
	if updated.HasActiveEnvironment() {
		if _, err := s.findEnvironment.Find(ctx, workspace.Path, id); err != nil {
			return "", err
		}
	}
	if err := s.saveLocalVariables.Save(ctx, localId, updated); err != nil {
		return "", err
	}
	if err := s.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{}, ""); err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": updated.ActiveEnvironmentID})
	return updated.ActiveEnvironmentID, nil
}
