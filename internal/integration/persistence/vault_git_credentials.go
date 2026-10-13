package persistence

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	gitKeychainService  = "dev.brienze.xenon.git"
	gitDataDir          = "git"
	gitCredentialsDir   = "credentials"
	gitCredentialSuffix = ".age"
)

func (v *vaultPersistence) CheckCredentialStore(ctx context.Context) error {
	logger.Debug(ctx, "Started")

	settings, err := readVaultSettings()
	if err != nil {
		return err
	}
	if _, unlocked := v.passphrase(); settings.UsesPassphrase() && !unlocked {
		return errs.ConflictError("the vault is locked; choose or enter the vault passphrase before adding Git credentials", "ACC-16409")
	}

	logger.Debug(ctx, "Finished")
	return nil
}

func (v *vaultPersistence) SaveCredential(ctx context.Context, accountId, secret string) error {
	logger.Debug(ctx, "Started", map[string]any{"accountId": accountId})

	settings, err := readVaultSettings()
	if err != nil {
		return err
	}
	if !settings.UsesPassphrase() {
		if err := v.keychain.Set(gitKeychainService, accountId, secret); err != nil {
			return errs.ServiceUnavailableError(err, "ACC-25503", "the Git credential could not be stored in the OS keychain")
		}
		logger.Debug(ctx, "Finished", map[string]any{"accountId": accountId, "store": "keychain"})
		return nil
	}
	passphrase, unlocked := v.passphrase()
	if !unlocked {
		return errs.ConflictError("the vault is locked; enter the vault passphrase first", "ACC-26409")
	}
	sealed, err := agecrypt.EncryptWithPassphrase([]byte(secret), passphrase, properties.Properties().Vault.ScryptLogN)
	if err != nil {
		return errs.InternalServerError(err, "ACC-27500", "the Git credential could not be encrypted")
	}
	path := gitCredentialPath(accountId)
	if err := writeVaultFile(path, sealed); err != nil {
		return err
	}
	v.mu.Lock()
	v.identities[path] = secret
	v.mu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"accountId": accountId, "store": "passphrase"})
	return nil
}

func (v *vaultPersistence) FindCredential(ctx context.Context, accountId string) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"accountId": accountId})

	settings, err := readVaultSettings()
	if err != nil {
		return "", err
	}
	if !settings.UsesPassphrase() {
		secret, _, err := v.keychain.Get(gitKeychainService, accountId)
		if err != nil {
			return "", errs.ServiceUnavailableError(err, "ACC-28503", "the Git credential could not be read from the OS keychain")
		}
		logger.Debug(ctx, "Finished", map[string]any{"accountId": accountId, "found": secret != ""})
		return secret, nil
	}
	secret, err := v.sealedCredential(gitCredentialPath(accountId))
	if err != nil {
		return "", err
	}

	logger.Debug(ctx, "Finished", map[string]any{"accountId": accountId, "found": secret != ""})
	return secret, nil
}

func (v *vaultPersistence) DeleteCredential(ctx context.Context, accountId string) error {
	logger.Debug(ctx, "Started", map[string]any{"accountId": accountId})

	path := gitCredentialPath(accountId)
	v.mu.Lock()
	delete(v.identities, path)
	v.mu.Unlock()
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.FileError(err, "ACC-29500")
	}
	settings, err := readVaultSettings()
	if err != nil {
		return err
	}
	if settings.UsesKeychain() {
		if err := v.keychain.Delete(gitKeychainService, accountId); err != nil {
			return errs.ServiceUnavailableError(err, "ACC-30503", "the Git credential could not be removed from the OS keychain")
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"accountId": accountId})
	return nil
}

func (v *vaultPersistence) sealedCredential(path string) (string, error) {
	if secret, cached := v.cachedIdentity(path); cached {
		return secret, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errs.FileError(err, "ACC-31500")
	}
	passphrase, unlocked := v.passphrase()
	if !unlocked {
		return "", errs.ConflictError("the vault is locked; enter the vault passphrase to use this Git account", "ACC-32409")
	}
	secret, err := agecrypt.DecryptWithPassphrase(data, passphrase)
	if err != nil {
		return "", errs.InternalServerError(err, "ACC-33500", "the Git credential could not be decrypted")
	}
	v.mu.Lock()
	v.identities[path] = string(secret)
	v.mu.Unlock()
	return string(secret), nil
}

func gitCredentialPath(accountId string) string {
	return appDataPath(gitDataDir, gitCredentialsDir, accountId+gitCredentialSuffix)
}
