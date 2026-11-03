package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAISession struct {
	findWorkspaceRegistry  usecase.FindWorkspaceRegistry
	loadWorkspace          usecase.LoadWorkspace
	findLocalVariables     usecase.FindLocalVariables
	findEnvironment        usecase.FindEnvironment
	listEnvironments       usecase.ListEnvironments
	findAISessionContext   usecase.FindAISessionContext
	saveAISessionContext   usecase.SaveAISessionContext
	deleteAISessionContext usecase.DeleteAISessionContext
}

func NewManageAISessionService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadWorkspace usecase.LoadWorkspace,
	findLocalVariables usecase.FindLocalVariables,
	findEnvironment usecase.FindEnvironment,
	listEnvironments usecase.ListEnvironments,
	findAISessionContext usecase.FindAISessionContext,
	saveAISessionContext usecase.SaveAISessionContext,
	deleteAISessionContext usecase.DeleteAISessionContext,
) adapter.ManageAISessionService {
	return &manageAISession{
		findWorkspaceRegistry:  findWorkspaceRegistry,
		loadWorkspace:          loadWorkspace,
		findLocalVariables:     findLocalVariables,
		findEnvironment:        findEnvironment,
		listEnvironments:       listEnvironments,
		findAISessionContext:   findAISessionContext,
		saveAISessionContext:   saveAISessionContext,
		deleteAISessionContext: deleteAISessionContext,
	}
}

func (m *manageAISession) Open(ctx context.Context, sessionId string) (*entity.AISessionContext, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	opened, err := m.current(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId, "result": *opened})
	return opened, nil
}

func (m *manageAISession) Current(ctx context.Context, sessionId string) (*entity.AISessionContext, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	current, err := m.current(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId, "result": *current})
	return current, nil
}

func (m *manageAISession) Enter(ctx context.Context, sessionId, localId string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId, "localId": localId})

	registered, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	if _, err := m.loadWorkspace.Load(ctx, registered.Path); err != nil {
		return nil, err
	}
	current, err := m.current(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	storedActiveEnvironmentId := ""
	if !current.HasEntered(localId) {
		if storedActiveEnvironmentId, err = m.storedActiveEnvironmentID(ctx, localId); err != nil {
			return nil, err
		}
	}
	entered := current.Enter(localId, storedActiveEnvironmentId)
	m.saveAISessionContext.Save(ctx, entered)

	logger.Info(ctx, "Finished", entered, map[string]any{"sessionId": sessionId, "localId": localId, "result": registered})
	return registered, nil
}

func (m *manageAISession) ChooseEnvironment(ctx context.Context, sessionId, environmentId string) (*entity.AISessionContext, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId, "environmentId": environmentId})

	current, err := m.current(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, current.LocalID)
	if err != nil {
		return nil, err
	}
	chosen := current.ChooseEnvironment(environmentId)
	if chosen.HasEnvironment() {
		if _, err := m.findEnvironment.Find(ctx, workspace.Path, chosen.EnvironmentID()); err != nil {
			return nil, err
		}
	}
	m.saveAISessionContext.Save(ctx, chosen)

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId, "environmentId": environmentId, "result": chosen})
	return &chosen, nil
}

func (m *manageAISession) ResolveEnvironment(ctx context.Context, localId, ref string) (string, bool, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "ref": ref})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", false, err
	}
	environments, err := m.listEnvironments.List(ctx, workspace.Path)
	if err != nil {
		return "", false, err
	}
	environmentId, known := entity.ResolveEnvironmentRef(environments, ref)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "ref": ref, "result": environmentId, "known": known})
	return environmentId, known, nil
}

func (m *manageAISession) Close(ctx context.Context, sessionId string) error {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	m.deleteAISessionContext.Delete(ctx, sessionId)

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId})
	return nil
}

func (m *manageAISession) current(ctx context.Context, sessionId string) (*entity.AISessionContext, error) {
	registry, err := m.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	stored, found := m.findAISessionContext.Find(ctx, sessionId)
	if found {
		current := stored.InRegistry(registry)
		m.saveAISessionContext.Save(ctx, current)
		return &current, nil
	}
	seeded := entity.NewAISessionContext(sessionId, "", "")
	if active, found := registry.Active(); found {
		storedActiveEnvironmentId, err := m.storedActiveEnvironmentID(ctx, active.LocalID)
		if err != nil {
			return nil, err
		}
		seeded = seeded.Enter(active.LocalID, storedActiveEnvironmentId)
	}
	m.saveAISessionContext.Save(ctx, seeded)
	return &seeded, nil
}

func (m *manageAISession) storedActiveEnvironmentID(ctx context.Context, localId string) (string, error) {
	local, err := m.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return "", err
	}
	return local.ActiveEnvironmentID, nil
}
