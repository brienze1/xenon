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
	findEnvironment       usecase.FindEnvironment
	saveEnvironment       usecase.SaveEnvironment
	deleteEnvironment     usecase.DeleteEnvironment
	orders                environmentOrders
	values                scopeValues
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewManageEnvironmentService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	listEnvironments usecase.ListEnvironments,
	findEnvironment usecase.FindEnvironment,
	saveEnvironment usecase.SaveEnvironment,
	deleteEnvironment usecase.DeleteEnvironment,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageEnvironmentService {
	return &manageEnvironment{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findEnvironment:       findEnvironment,
		saveEnvironment:       saveEnvironment,
		deleteEnvironment:     deleteEnvironment,
		orders: environmentOrders{
			listEnvironments:  listEnvironments,
			loadWorkspace:     loadWorkspace,
			saveWorkspaceFile: saveWorkspaceFile,
		},
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
	list, err := m.orders.list(ctx, workspace.Path, local.ActiveEnvironmentID)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": *list})
	return list, nil
}

func (m *manageEnvironment) Get(ctx context.Context, localId, id string) (*entity.EnvironmentDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	_, environment, err := m.environment(ctx, localId, id)
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
	list, err := m.orders.list(ctx, workspace.Path, "")
	if err != nil {
		return nil, err
	}
	created := entity.NewEnvironment(idutils.NewUlid(), name)
	saved, err := m.saveNamed(ctx, workspace.Path, created)
	if err != nil {
		return nil, err
	}
	if err := m.orders.write(ctx, workspace.Path, list.OrderAppending(saved.ID)); err != nil {
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

	root, current, err := m.environment(ctx, localId, settings.ID)
	if err != nil {
		return nil, err
	}
	saved, err := m.saveNamed(ctx, root, current.WithSettings(settings))
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

	root, source, err := m.environment(ctx, localId, id)
	if err != nil {
		return nil, err
	}
	list, err := m.orders.list(ctx, root, "")
	if err != nil {
		return nil, err
	}
	saved, err := m.saveNamed(ctx, root, source.Duplicated(idutils.NewUlid(), list.CopyName(source.Name)))
	if err != nil {
		return nil, err
	}
	if err := m.orders.write(ctx, root, list.OrderAppending(saved.ID)); err != nil {
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

	root, environment, err := m.environment(ctx, localId, id)
	if err != nil {
		return "", err
	}
	listed, err := m.orders.lists(ctx, root, id)
	if err != nil {
		return "", err
	}
	list, err := m.orders.list(ctx, root, "")
	if err != nil {
		return "", err
	}
	if err := m.deleteEnvironment.Delete(ctx, root, id); err != nil {
		return "", err
	}
	if listed {
		if err := m.orders.write(ctx, root, list.OrderRemoving(id)); err != nil {
			return "", err
		}
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

func (m *manageEnvironment) environment(ctx context.Context, localId, id string) (string, *entity.Environment, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", nil, err
	}
	environment, err := m.findEnvironment.Find(ctx, workspace.Path, id)
	if err != nil {
		return "", nil, err
	}
	return workspace.Path, environment, nil
}

func (m *manageEnvironment) saveNamed(ctx context.Context, root string, environment entity.Environment) (*entity.Environment, error) {
	list, err := m.orders.list(ctx, root, "")
	if err != nil {
		return nil, err
	}
	if list.KeepsName(environment) {
		return m.saveEnvironment.Save(ctx, root, environment)
	}
	if !environment.HasValidName() {
		return nil, errs.BadRequestError("the environment name must have 1 to 200 characters", "ENV-16400")
	}
	if !list.IsNameAvailableFor(environment) {
		return nil, errs.ConflictError("an environment with this name already exists", "ENV-02409")
	}
	named := environment.WithFileName(usecase.DeriveFileName(environment.Name, itemkind.Environment, list.SiblingFileNames(environment)))
	return m.saveEnvironment.Save(ctx, root, named)
}

func (m *manageEnvironment) announced(ctx context.Context, localId string, environment entity.Environment) (*entity.EnvironmentDetails, error) {
	if err := m.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{environment.ScopeRef()}, ""); err != nil {
		return nil, err
	}
	return m.values.details(ctx, localId, environment)
}
