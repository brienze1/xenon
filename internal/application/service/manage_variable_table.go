package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageVariableTable struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	scopes                variableScopes
	rowStore              variableRowStore
	findLocalVariables    usecase.FindLocalVariables
	saveLocalVariables    usecase.SaveLocalVariables
	vaults                vaultWriter
	publishVariableEvent  usecase.PublishVariableEvent
}

type loadedTable struct {
	local entity.LocalVariables
	vault entity.Vault
	table entity.VariableTable
	root  string
}

func NewManageVariableTableService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findEnvironment usecase.FindEnvironment,
	saveEnvironment usecase.SaveEnvironment,
	findTreeItem usecase.FindTreeItem,
	saveTreeItem usecase.SaveTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageVariableTableService {
	scopes := variableScopes{findEnvironment: findEnvironment, findTreeItem: findTreeItem, loadWorkspace: loadWorkspace}
	return &manageVariableTable{
		findWorkspaceRegistry: findWorkspaceRegistry,
		scopes:                scopes,
		rowStore: variableRowStore{
			scopes:            scopes,
			saveEnvironment:   saveEnvironment,
			saveTreeItem:      saveTreeItem,
			saveWorkspaceFile: saveWorkspaceFile,
		},
		findLocalVariables: findLocalVariables,
		saveLocalVariables: saveLocalVariables,
		vaults: vaultWriter{
			findVault:            findVault,
			saveVault:            saveVault,
			findVaultSettings:    findVaultSettings,
			saveVaultSettings:    saveVaultSettings,
			probeKeychain:        probeKeychain,
			publishVariableEvent: publishVariableEvent,
		},
		publishVariableEvent: publishVariableEvent,
	}
}

func (m *manageVariableTable) GetTable(ctx context.Context, localId string, scope entity.ScopeRef) (*entity.VariableTable, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String()})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "result": loaded.table})
	return &loaded.table, nil
}

func (m *manageVariableTable) SaveTable(ctx context.Context, localId string, scope entity.ScopeRef, edits []entity.VariableEdit, confirmReveal bool) (*entity.VariableTable, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "edits": len(edits), "confirmReveal": confirmReveal})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}
	if loaded.table.RevealedBy(edits) && !confirmReveal {
		return nil, errs.ConflictError("turning a secret into a plain variable writes its value to the workspace files; confirm it", "VAR-01409")
	}
	edited := loaded.table.Edited(edits)
	if !edited.HasValidKeys() {
		return nil, errs.BadRequestError("variable names must not be empty or contain {, } or line breaks", "VAR-00400")
	}
	saved, err := m.store(ctx, localId, *loaded, edited, true)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "result": *saved})
	return saved, nil
}

func (m *manageVariableTable) SetCurrent(ctx context.Context, localId string, scope entity.ScopeRef, key, value string) (*entity.VariableTableEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "key": key})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}
	if !loaded.table.Contains(key) {
		return nil, errs.NotFoundError("the table has no variable with this name", "VAR-03404")
	}
	saved, err := m.store(ctx, localId, *loaded, loaded.table.WithCurrent(key, value), false)
	if err != nil {
		return nil, err
	}
	entry, _ := saved.Entry(key)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "key": key, "result": entry})
	return &entry, nil
}

func (m *manageVariableTable) Reset(ctx context.Context, localId string, scope entity.ScopeRef, keys []string) (*entity.VariableTable, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "keys": keys})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}
	saved, err := m.store(ctx, localId, *loaded, loaded.table.Reset(keys), false)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "keys": keys, "result": *saved})
	return saved, nil
}

func (m *manageVariableTable) Persist(ctx context.Context, localId string, scope entity.ScopeRef, keys []string) (*entity.VariableTable, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "keys": keys})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}
	if loaded.table.PersistsSecret(keys) {
		return nil, errs.UnprocessableEntityError("secret variables have no initial value to persist", "VAR-02422")
	}
	saved, err := m.store(ctx, localId, *loaded, loaded.table.Persisted(keys), true)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "keys": keys, "result": *saved})
	return saved, nil
}

func (m *manageVariableTable) AddVariable(ctx context.Context, localId string, scope entity.ScopeRef, row entity.VariableRow, value string) (*entity.VariableTableEntry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "row": row})

	loaded, err := m.load(ctx, localId, scope)
	if err != nil {
		return nil, err
	}
	if !row.HasValidKey() {
		return nil, errs.BadRequestError("variable names must not be empty or contain {, } or line breaks", "VAR-13400")
	}
	if loaded.table.Contains(row.Key) {
		return nil, errs.ConflictError("the table already has a variable with this name", "VAR-04409")
	}
	saved, err := m.store(ctx, localId, *loaded, loaded.table.Added(row, value), true)
	if err != nil {
		return nil, err
	}
	entry, _ := saved.Entry(row.Key)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "row": row, "result": entry})
	return &entry, nil
}

func (m *manageVariableTable) load(ctx context.Context, localId string, scope entity.ScopeRef) (*loadedTable, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	rows, err := m.scopes.rows(ctx, workspace.Path, scope)
	if err != nil {
		return nil, err
	}
	local, err := m.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return nil, err
	}
	vault, err := vaultFor(ctx, m.vaults.findVault, localId, rows)
	if err != nil {
		return nil, err
	}
	return &loadedTable{root: workspace.Path, local: *local, vault: *vault, table: entity.NewVariableTable(scope, rows, *local, *vault)}, nil
}

// store writes the vault, then the local current values, then the file, restoring the earlier stores when a later write fails.
func (m *manageVariableTable) store(ctx context.Context, localId string, loaded loadedTable, edited entity.VariableTable, rowsChanged bool) (*entity.VariableTable, error) {
	vault, err := m.storeSecrets(ctx, loaded, edited)
	if err != nil {
		return nil, err
	}
	restoreVault := func() {
		if edited.SecretsDiffer(loaded.table) {
			m.vaults.restore(ctx, vault, edited.Scope, loaded.table.Secrets)
		}
	}
	if err := m.saveLocalVariables.Save(ctx, localId, loaded.local.WithCurrentValues(edited.Scope, edited.StoredCurrent())); err != nil {
		restoreVault()
		return nil, err
	}
	if rowsChanged {
		if err := m.rowStore.save(ctx, loaded.root, edited.Scope, edited.Rows); err != nil {
			m.restoreLocal(ctx, localId, loaded.local)
			restoreVault()
			return nil, err
		}
	}
	if err := m.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{edited.Scope}, ""); err != nil {
		return nil, err
	}
	stored := entity.NewVariableTable(edited.Scope, edited.Rows, loaded.local.WithCurrentValues(edited.Scope, edited.StoredCurrent()), vault)
	stored.Warnings = edited.Warnings
	return &stored, nil
}

func (m *manageVariableTable) storeSecrets(ctx context.Context, loaded loadedTable, edited entity.VariableTable) (entity.Vault, error) {
	if !edited.SecretsDiffer(loaded.table) {
		return loaded.vault, nil
	}
	return m.vaults.write(ctx, loaded.vault, edited.Scope, edited.StoredSecrets())
}

func (m *manageVariableTable) restoreLocal(ctx context.Context, localId string, local entity.LocalVariables) {
	if err := m.saveLocalVariables.Save(ctx, localId, local); err != nil {
		logger.Error(ctx, err, "The local current values could not be restored after a failed save", map[string]any{"localId": localId})
	}
}
