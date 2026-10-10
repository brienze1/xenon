package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type scopeValues struct {
	findLocalVariables usecase.FindLocalVariables
	saveLocalVariables usecase.SaveLocalVariables
	findVault          usecase.FindVault
	saveVault          usecase.SaveVault
}

func (s scopeValues) local(ctx context.Context, localId string) (*entity.LocalVariables, error) {
	local, err := s.findLocalVariables.Find(ctx, localId)
	return localVariablesOrDefaults(ctx, localId, local, err)
}

func (s scopeValues) copy(ctx context.Context, localId string, source, target entity.ScopeRef, rows []entity.VariableRow) error {
	local, err := s.local(ctx, localId)
	if err != nil {
		return err
	}
	if err := s.saveLocalVariables.Save(ctx, localId, local.WithCurrentValues(target, local.CurrentValues(source))); err != nil {
		return err
	}
	vault, err := vaultFor(ctx, s.findVault, localId, rows)
	if err != nil || !vault.HasScope(source) {
		return err
	}
	return s.saveVault.Save(ctx, vault.WithSecrets(target, vault.SecretsFor(source)))
}

func (s scopeValues) drop(ctx context.Context, localId string, environment entity.Environment) (string, error) {
	local, err := s.local(ctx, localId)
	if err != nil {
		return "", err
	}
	updated := local.WithCurrentValues(environment.ScopeRef(), nil)
	if local.IsActive(environment.ID) {
		updated = updated.WithActiveEnvironment("")
	}
	if err := s.saveLocalVariables.Save(ctx, localId, updated); err != nil {
		return "", err
	}
	vault, err := vaultFor(ctx, s.findVault, localId, environment.Variables)
	if err != nil || !vault.HasScope(environment.ScopeRef()) {
		return updated.ActiveEnvironmentID, err
	}
	return updated.ActiveEnvironmentID, s.saveVault.Save(ctx, vault.WithoutScope(environment.ScopeRef()))
}

func (s scopeValues) details(ctx context.Context, localId string, environment entity.Environment) (*entity.EnvironmentDetails, error) {
	local, err := s.local(ctx, localId)
	if err != nil {
		return nil, err
	}
	vault, err := vaultFor(ctx, s.findVault, localId, environment.Variables)
	if err != nil {
		return nil, err
	}
	details := entity.NewEnvironmentDetails(environment, *local, *vault)
	return &details, nil
}
