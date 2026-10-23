package service

import (
	"context"
	"maps"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

const scriptChangeReason = "script"

type scriptVariables struct {
	scopes               variableScopes
	rowStore             variableRowStore
	findLocalVariables   usecase.FindLocalVariables
	saveLocalVariables   usecase.SaveLocalVariables
	vaults               vaultWriter
	publishVariableEvent usecase.PublishVariableEvent
	lockLocalVariables   usecase.LockLocalVariables
}

func (s scriptVariables) persist(ctx context.Context, localId, root string, refs map[enums.VariableScope]entity.ScopeRef, before, after entity.ScriptScopes, vaultBefore, vaultAfter map[string]string) error {
	unlock := s.lockLocalVariables.LockVariables(ctx, localId)
	defer unlock()
	changes := after.ChangesFrom(before, refs, vaultBefore, vaultAfter)
	changed := make([]entity.ScopeRef, 0, len(changes))
	for _, change := range changes {
		if err := s.apply(ctx, localId, root, change); err != nil {
			return err
		}
		changed = entity.AppendScopeRef(changed, change.Ref)
	}
	if len(changed) == 0 {
		return nil
	}
	logger.Info(ctx, "Script variable changes saved", map[string]any{"localId": localId, "scopes": len(changed)})
	return s.publishVariableEvent.VariablesChanged(ctx, localId, changed, scriptChangeReason)
}

func (s scriptVariables) apply(ctx context.Context, localId, root string, change entity.ScriptScopeChange) error {
	ref := change.Ref
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
	edited := original.WithScriptChanges(change.Set, change.Unset, change.Secret)
	if edited.SecretsDiffer(original) {
		if _, err := s.vaults.write(ctx, *vault, ref, edited.StoredSecrets()); err != nil {
			return err
		}
	}
	if err := s.saveLocalVariables.Save(ctx, localId, local.WithCurrentValues(ref, edited.StoredCurrent())); err != nil {
		return err
	}
	if !edited.RowsDiffer(original) {
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
