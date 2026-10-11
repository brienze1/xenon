package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestVariables struct {
	findTreeItem       usecase.FindTreeItem
	loadWorkspace      usecase.LoadWorkspace
	findEnvironment    usecase.FindEnvironment
	findLocalVariables usecase.FindLocalVariables
	findVault          usecase.FindVault
}

type requestScopes struct {
	environment *entity.Environment
	vault       map[string]string
	ancestors   []entity.Collection
	context     entity.VariableContext
	scripts     entity.ScriptScopes
}

func (r requestVariables) load(ctx context.Context, localId, root, requestId string) (*requestScopes, error) {
	ancestors := make([]entity.Collection, 0)
	if entity.IsSavedRequestID(requestId) {
		found, err := r.findTreeItem.FindAncestors(ctx, root, requestId)
		if err != nil {
			return nil, err
		}
		ancestors = found
	}
	local, err := r.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return nil, err
	}
	environment, err := r.activeEnvironment(ctx, root, *local)
	if err != nil {
		return nil, err
	}
	workspace, err := r.loadWorkspace.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	vault, err := r.vault(ctx, localId, entity.UsesSecretRows(environment, ancestors, workspace.Variables))
	if err != nil {
		return nil, err
	}
	return &requestScopes{
		environment: environment,
		ancestors:   ancestors,
		context:     usecase.BuildVariableContext(environment, ancestors, workspace.Variables, *local, *vault),
		scripts:     usecase.BuildScriptScopes(environment, ancestors, workspace.Variables, *local, *vault),
		vault:       usecase.BuildScriptVault(workspace.Variables, *vault),
	}, nil
}

func (r requestVariables) vault(ctx context.Context, localId string, needed bool) (*entity.Vault, error) {
	if !needed {
		vault := entity.NewUnopenedVault(localId)
		return &vault, nil
	}
	return r.findVault.Find(ctx, localId)
}

func (r requestVariables) activeEnvironment(ctx context.Context, root string, local entity.LocalVariables) (*entity.Environment, error) {
	if !local.HasActiveEnvironment() {
		return nil, nil
	}
	environment, err := r.findEnvironment.Find(ctx, root, local.ActiveEnvironmentID)
	if errs.IsNotFoundError(err) {
		logger.Info(ctx, "The active environment no longer exists, resolving without one", local)
		return nil, nil
	}
	return environment, err
}
