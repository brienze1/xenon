package service

import (
	"context"
	"maps"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
	"github.com/brienze1/xenon/pkg/logger"
)

const scriptChangeReason = "script"

var persistedScopes = []enums.VariableScope{variablescope.Environment, variablescope.Collection, variablescope.Globals}

type scriptVariables struct {
	scopes               variableScopes
	rowStore             variableRowStore
	findLocalVariables   usecase.FindLocalVariables
	saveLocalVariables   usecase.SaveLocalVariables
	vaults               vaultWriter
	publishVariableEvent usecase.PublishVariableEvent
}

type scriptChange struct {
	set    []entity.KeyValue
	unset  []string
	secret bool
}

func (s scriptVariables) persist(ctx context.Context, localId, root string, refs map[enums.VariableScope]entity.ScopeRef, before, after entity.ScriptScopes, vaultBefore, vaultAfter map[string]string) error {
	changed := make([]entity.ScopeRef, 0, len(persistedScopes))
	for _, kind := range persistedScopes {
		ref, found := refs[kind]
		set, unset := after.Scope(kind).ChangesFrom(before.Scope(kind))
		if !found || (len(set) == 0 && len(unset) == 0) {
			continue
		}
		if err := s.apply(ctx, localId, root, ref, scriptChange{set: set, unset: unset}); err != nil {
			return err
		}
		changed = entity.AppendScopeRef(changed, ref)
	}
	if set, unset := entity.VaultChangesFrom(vaultBefore, vaultAfter); len(set) > 0 || len(unset) > 0 {
		if err := s.apply(ctx, localId, root, entity.NewGlobalsScopeRef(), scriptChange{set: set, unset: unset, secret: true}); err != nil {
			return err
		}
		changed = entity.AppendScopeRef(changed, entity.NewGlobalsScopeRef())
	}
	if len(changed) == 0 {
		return nil
	}
	logger.Info(ctx, "Script variable changes saved", map[string]any{"localId": localId, "scopes": len(changed)})
	return s.publishVariableEvent.VariablesChanged(ctx, localId, changed, scriptChangeReason)
}

func (s scriptVariables) apply(ctx context.Context, localId, root string, ref entity.ScopeRef, change scriptChange) error {
	rows, err := s.scopes.rows(ctx, root, ref)
	if err != nil {
		return err
	}
	local, err := s.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return err
	}
	vault, err := vaultFor(ctx, s.vaults.findVault, localId, rows)
	if err != nil {
		return err
	}
	original := entity.NewVariableTable(ref, rows, *local, *vault)
	edited := original.WithScriptChanges(change.set, change.unset, change.secret)
	if edited.SecretsDiffer(original) {
		if _, err := s.vaults.write(ctx, *vault, ref, edited.StoredSecrets()); err != nil {
			return err
		}
	}
	if err := s.saveLocalVariables.Save(ctx, localId, local.WithCurrentValues(ref, edited.StoredCurrent())); err != nil {
		return err
	}
	if len(edited.Rows) == len(original.Rows) {
		return nil
	}
	return s.rowStore.save(ctx, root, ref, edited.Rows)
}

func cloneVault(vault map[string]string) map[string]string {
	if vault == nil {
		return map[string]string{}
	}
	return maps.Clone(vault)
}
