package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageEnvironment struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	listEnvironments      usecase.ListEnvironments
	findEnvironment       usecase.FindEnvironment
	saveEnvironment       usecase.SaveEnvironment
	deleteEnvironment     usecase.DeleteEnvironment
	saveLocalVariables    usecase.SaveLocalVariables
	values                scopeValues
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewManageEnvironmentService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	listEnvironments usecase.ListEnvironments,
	findEnvironment usecase.FindEnvironment,
	saveEnvironment usecase.SaveEnvironment,
	deleteEnvironment usecase.DeleteEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageEnvironmentService {
	return &manageEnvironment{
		findWorkspaceRegistry: findWorkspaceRegistry,
		listEnvironments:      listEnvironments,
		findEnvironment:       findEnvironment,
		saveEnvironment:       saveEnvironment,
		deleteEnvironment:     deleteEnvironment,
		saveLocalVariables:    saveLocalVariables,
		values: scopeValues{
			findLocalVariables: findLocalVariables,
			saveLocalVariables: saveLocalVariables,
			findVault:          findVault,
			saveVault:          saveVault,
		},
		publishVariableEvent: publishVariableEvent,
	}
}

func (m *manageEnvironment) List(ctx context.Context, localId string) (*entity.EnvironmentList, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	local, err := m.values.local(ctx, localId)
	if err != nil {
		return nil, err
	}
	list, err := m.list(ctx, workspace.Path, local.ActiveEnvironmentID)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": *list})
	return list, nil
}

func (m *manageEnvironment) SetActive(ctx context.Context, localId, id string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", err
	}
	local, err := m.values.local(ctx, localId)
	if err != nil {
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
	if err := m.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{}, ""); err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": updated.ActiveEnvironmentID})
	return updated.ActiveEnvironmentID, nil
}

func (m *manageEnvironment) Get(ctx context.Context, localId, id string) (*entity.EnvironmentDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	environment, err := m.findEnvironment.Find(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}
	details, err := m.values.details(ctx, localId, *environment)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *details})
	return details, nil
}

func (m *manageEnvironment) Create(ctx context.Context, localId, name string) (*entity.EnvironmentDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "name": name})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	created := entity.NewEnvironment(idutils.NewUlid(), name)
	saved, err := m.saveNamed(ctx, workspace.Path, created, "")
	if err != nil {
		return nil, err
	}
	details, err := m.announced(ctx, localId, *saved)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "name": name, "result": *details})
	return details, nil
}

func (m *manageEnvironment) Update(ctx context.Context, localId string, settings entity.Environment) (*entity.EnvironmentDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "settings": settings})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	current, err := m.findEnvironment.Find(ctx, workspace.Path, settings.ID)
	if err != nil {
		return nil, err
	}
	updated := current.WithSettings(settings)
	saved, err := m.saveRenamed(ctx, workspace.Path, *current, updated)
	if err != nil {
		return nil, err
	}
	details, err := m.announced(ctx, localId, *saved)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "settings": settings, "result": *details})
	return details, nil
}

func (m *manageEnvironment) Duplicate(ctx context.Context, localId, id string) (*entity.EnvironmentDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	source, err := m.findEnvironment.Find(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}
	list, err := m.list(ctx, workspace.Path, "")
	if err != nil {
		return nil, err
	}
	saved, err := m.saveNamed(ctx, workspace.Path, source.Duplicated(idutils.NewUlid(), list.CopyName(source.Name)), "")
	if err != nil {
		return nil, err
	}
	if err := m.values.copy(ctx, localId, source.ScopeRef(), saved.ScopeRef(), source.Variables); err != nil {
		return nil, err
	}
	details, err := m.announced(ctx, localId, *saved)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *details})
	return details, nil
}

func (m *manageEnvironment) Delete(ctx context.Context, localId, id string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", err
	}
	environment, err := m.findEnvironment.Find(ctx, workspace.Path, id)
	if err != nil {
		return "", err
	}
	if err := m.deleteEnvironment.Delete(ctx, workspace.Path, id); err != nil {
		return "", err
	}
	activeId, err := m.values.drop(ctx, localId, *environment)
	if err != nil {
		return "", err
	}
	if err := m.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{environment.ScopeRef()}, ""); err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": activeId})
	return activeId, nil
}

func (m *manageEnvironment) list(ctx context.Context, root, activeId string) (*entity.EnvironmentList, error) {
	environments, err := m.listEnvironments.List(ctx, root)
	if err != nil {
		return nil, err
	}
	list := entity.NewEnvironmentList(environments, activeId)
	return &list, nil
}

func (m *manageEnvironment) saveNamed(ctx context.Context, root string, environment entity.Environment, exceptId string) (*entity.Environment, error) {
	if !environment.HasValidName() {
		return nil, errs.BadRequestError("the environment name must have 1 to 200 characters", "ENV-16400")
	}
	list, err := m.list(ctx, root, "")
	if err != nil {
		return nil, err
	}
	if list.IsNameTaken(environment.Name, exceptId) {
		return nil, errs.ConflictError("an environment with this name already exists", "ENV-02409")
	}
	named := environment.WithFileName(usecase.DeriveFileName(environment.Name, itemkind.Environment, list.FileNames(exceptId)))
	return m.saveEnvironment.Save(ctx, root, named)
}

func (m *manageEnvironment) saveRenamed(ctx context.Context, root string, current, updated entity.Environment) (*entity.Environment, error) {
	if current.HasName(updated.Name) {
		return m.saveEnvironment.Save(ctx, root, updated)
	}
	return m.saveNamed(ctx, root, updated, current.ID)
}

func (m *manageEnvironment) announced(ctx context.Context, localId string, environment entity.Environment) (*entity.EnvironmentDetails, error) {
	if err := m.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{environment.ScopeRef()}, ""); err != nil {
		return nil, err
	}
	return m.values.details(ctx, localId, environment)
}
