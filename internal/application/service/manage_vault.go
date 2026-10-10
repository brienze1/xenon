package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageVault struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	scopes                variableScopes
	findVault             usecase.FindVault
	deleteVault           usecase.DeleteVault
	findVaultSettings     usecase.FindVaultSettings
	saveVaultSettings     usecase.SaveVaultSettings
	sealVaultPassphrase   usecase.SealVaultPassphrase
	verifyVaultPassphrase usecase.VerifyVaultPassphrase
	unlockVaults          usecase.UnlockVaults
	copyToClipboard       usecase.CopyToClipboard
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewManageVaultService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findEnvironment usecase.FindEnvironment,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findVault usecase.FindVault,
	deleteVault usecase.DeleteVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	sealVaultPassphrase usecase.SealVaultPassphrase,
	verifyVaultPassphrase usecase.VerifyVaultPassphrase,
	unlockVaults usecase.UnlockVaults,
	copyToClipboard usecase.CopyToClipboard,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageVaultService {
	return &manageVault{
		findWorkspaceRegistry: findWorkspaceRegistry,
		scopes:                variableScopes{findEnvironment: findEnvironment, findTreeItem: findTreeItem, loadWorkspace: loadWorkspace},
		findVault:             findVault,
		deleteVault:           deleteVault,
		findVaultSettings:     findVaultSettings,
		saveVaultSettings:     saveVaultSettings,
		sealVaultPassphrase:   sealVaultPassphrase,
		verifyVaultPassphrase: verifyVaultPassphrase,
		unlockVaults:          unlockVaults,
		copyToClipboard:       copyToClipboard,
		publishVariableEvent:  publishVariableEvent,
	}
}

func (m *manageVault) Status(ctx context.Context) (*entity.VaultStatus, error) {
	logger.Info(ctx, "Started")

	status, err := m.status(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"result": *status})
	return status, nil
}

func (m *manageVault) SetupPassphrase(ctx context.Context, setup entity.PassphraseSetup) (*entity.VaultStatus, error) {
	logger.Info(ctx, "Started", setup)

	if !setup.IsConfirmed() {
		return nil, errs.BadRequestError("the passphrase and its confirmation differ", "VLT-00400")
	}
	if !setup.IsLongEnough() {
		return nil, errs.BadRequestError("the passphrase must have at least 8 characters", "VLT-01400")
	}
	settings, err := m.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings.HasPassphrase() {
		return nil, errs.ConflictError("a vault passphrase was already chosen", "VLT-09409")
	}
	check, err := m.sealVaultPassphrase.Seal(ctx, setup.Passphrase)
	if err != nil {
		return nil, err
	}
	if err := m.saveVaultSettings.SaveSettings(ctx, settings.WithPassphraseCheck(check)); err != nil {
		return nil, err
	}
	status, err := m.unlocked(ctx, setup.Passphrase)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", setup, map[string]any{"result": *status})
	return status, nil
}

func (m *manageVault) Unlock(ctx context.Context, passphrase string) (*entity.VaultStatus, error) {
	logger.Info(ctx, "Started")

	settings, err := m.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.HasPassphrase() {
		return nil, errs.ConflictError("no vault passphrase was chosen yet", "VLT-08409")
	}
	if !m.verifyVaultPassphrase.Verify(ctx, settings.PassphraseCheck, passphrase) {
		return nil, errs.UnauthorizedError("the passphrase is wrong", "VLT-02401")
	}
	status, err := m.unlocked(ctx, passphrase)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"result": *status})
	return status, nil
}

func (m *manageVault) Reveal(ctx context.Context, localId string, scope entity.ScopeRef, key string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "key": key})

	value, err := m.secret(ctx, localId, scope, key)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "key": key})
	return value, nil
}

func (m *manageVault) CopySecret(ctx context.Context, localId string, scope entity.ScopeRef, key string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "scope": scope.String(), "key": key})

	value, err := m.secret(ctx, localId, scope, key)
	if err != nil {
		return err
	}
	if err := m.copyToClipboard.Copy(ctx, value); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "scope": scope.String(), "key": key})
	return nil
}

func (m *manageVault) Reset(ctx context.Context, localId string, confirm bool) (*entity.VaultStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "confirm": confirm})

	if !confirm {
		return nil, errs.BadRequestError("resetting the vault discards its secret values; confirm it", "VLT-07400")
	}
	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	if err := m.deleteVault.Delete(ctx, localId); err != nil {
		return nil, err
	}
	status, err := m.announced(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "confirm": confirm, "result": *status})
	return status, nil
}

func (m *manageVault) secret(ctx context.Context, localId string, scope entity.ScopeRef, key string) (string, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", err
	}
	rows, err := m.scopes.rows(ctx, workspace.Path, scope)
	if err != nil {
		return "", err
	}
	vault, err := m.findVault.Find(ctx, localId)
	if err != nil {
		return "", err
	}
	settings, err := m.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return "", err
	}
	if err := usecase.CheckVaultReadable(*vault, *settings); err != nil {
		return "", err
	}
	value, found := entity.NewVariableTable(scope, rows, entity.NewLocalVariables(), *vault).Secret(key)
	if !found {
		return "", errs.NotFoundError("the secret has no value on this machine", "VLT-03404")
	}
	return value, nil
}

func (m *manageVault) unlocked(ctx context.Context, passphrase string) (*entity.VaultStatus, error) {
	m.unlockVaults.Unlock(ctx, passphrase)
	return m.announced(ctx)
}

func (m *manageVault) announced(ctx context.Context) (*entity.VaultStatus, error) {
	status, err := m.status(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.publishVariableEvent.VaultStatusChanged(ctx, *status); err != nil {
		return nil, err
	}
	return status, nil
}

func (m *manageVault) status(ctx context.Context) (*entity.VaultStatus, error) {
	settings, err := m.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return nil, err
	}
	registry, err := m.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	vaults := make([]entity.Vault, 0, len(registry.Workspaces))
	for _, workspace := range registry.Workspaces {
		vault, err := m.findVault.Find(ctx, workspace.LocalID)
		if err != nil {
			return nil, err
		}
		vaults = append(vaults, *vault)
	}
	status := entity.NewVaultStatus(*settings, vaults)
	return &status, nil
}
