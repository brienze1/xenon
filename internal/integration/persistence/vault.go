package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	vaultFile            = "vault.age"
	vaultKeyFile         = "vault-key.age"
	vaultKeychainService = "dev.brienze.xenon.vault"
)

type vaultIdentity struct {
	secret       string
	found        bool
	sealed       bool
	unreadable   bool
	keyAvailable bool
}

type vaultPersistence struct {
	keychain    adapter.Keychain
	passphrases map[string]string
	identities  map[string]string
	mu          sync.RWMutex
}

func NewVaultPersistence(keychain adapter.Keychain) adapter.VaultPersistence {
	return &vaultPersistence{keychain: keychain, passphrases: map[string]string{}, identities: map[string]string{}}
}

func (v *vaultPersistence) Find(ctx context.Context, localId string) (*entity.Vault, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	settings, err := readVaultSettings()
	if err != nil {
		return nil, err
	}
	identity, err := v.identity(localId, *settings)
	if err != nil {
		return nil, err
	}
	vault := entity.NewVaultFromStorage(v.storage(localId, *settings, *identity))

	logger.Debug(ctx, "Finished", vault)
	return &vault, nil
}

func (v *vaultPersistence) Save(ctx context.Context, vault entity.Vault) error {
	logger.Debug(ctx, "Started", vault)

	settings, err := readVaultSettings()
	if err != nil {
		return err
	}
	identity, err := v.identity(vault.LocalID, *settings)
	if err != nil {
		return err
	}
	secret := identity.secret
	if !identity.found {
		if secret, err = v.createIdentity(vault.LocalID, *settings); err != nil {
			return err
		}
	}
	content, _ := json.Marshal(model.NewVaultContent(vault))
	encrypted, err := agecrypt.EncryptForIdentity(content, secret)
	if err != nil {
		return errs.InternalServerError(err, "VLT-20500", "the vault could not be encrypted")
	}
	if err := writeVaultFile(vaultPath(vault.LocalID), encrypted); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", vault)
	return nil
}

func (v *vaultPersistence) Delete(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	settings, err := readVaultSettings()
	if err != nil {
		return err
	}
	v.mu.Lock()
	delete(v.identities, vaultKeyPath(localId))
	v.mu.Unlock()
	for _, path := range []string{vaultPath(localId), vaultKeyPath(localId)} {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errs.FileError(err, "VLT-23500")
		}
	}
	if settings.Mode == vaultmode.Keychain {
		if err := v.keychain.Delete(vaultKeychainService, localId); err != nil {
			return errs.ServiceUnavailableError(err, "VLT-24503", "the vault key could not be removed from the OS keychain")
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (v *vaultPersistence) Unlock(ctx context.Context, passphrase string) {
	logger.Debug(ctx, "Started")

	v.mu.Lock()
	v.passphrases[appDataPath()] = passphrase
	v.mu.Unlock()

	logger.Debug(ctx, "Finished")
}

func (v *vaultPersistence) identity(localId string, settings entity.VaultSettings) (*vaultIdentity, error) {
	switch settings.Mode {
	case vaultmode.Keychain:
		secret, found, err := v.keychain.Get(vaultKeychainService, localId)
		if err != nil {
			return nil, errs.InternalServerError(err, "VLT-15500", "the vault key could not be read from the OS keychain")
		}
		return &vaultIdentity{secret: secret, found: found, keyAvailable: true}, nil
	case vaultmode.Passphrase:
		return v.wrappedIdentity(localId), nil
	default:
		return &vaultIdentity{}, nil
	}
}

func (v *vaultPersistence) wrappedIdentity(localId string) *vaultIdentity {
	path := vaultKeyPath(localId)
	passphrase, unlocked := v.passphrase()
	if !isFile(path) {
		return &vaultIdentity{keyAvailable: unlocked}
	}
	if secret, cached := v.cachedIdentity(path); cached {
		return &vaultIdentity{secret: secret, found: true, keyAvailable: unlocked}
	}
	if !unlocked {
		return &vaultIdentity{found: true, sealed: true}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &vaultIdentity{found: true, unreadable: true, keyAvailable: true}
	}
	secret, err := agecrypt.DecryptWithPassphrase(data, passphrase)
	if err != nil {
		return &vaultIdentity{found: true, unreadable: true, keyAvailable: true}
	}
	v.mu.Lock()
	v.identities[path] = string(secret)
	v.mu.Unlock()
	return &vaultIdentity{secret: string(secret), found: true, keyAvailable: true}
}

func (v *vaultPersistence) storage(localId string, settings entity.VaultSettings, identity vaultIdentity) entity.VaultStorage {
	path := vaultPath(localId)
	storage := entity.VaultStorage{
		LocalID:       localId,
		Mode:          settings.Mode,
		KeyFound:      identity.found,
		KeySealed:     identity.sealed,
		KeyUnreadable: identity.unreadable,
		KeyAvailable:  identity.keyAvailable,
		ContentFound:  isFile(path),
	}
	if !storage.CanOpenContent() {
		return storage
	}
	data, err := os.ReadFile(path)
	if err != nil {
		storage.ContentUnreadable = true
		return storage
	}
	plaintext, err := agecrypt.DecryptWithIdentity(data, identity.secret)
	var content model.VaultContent
	if err != nil || json.Unmarshal(plaintext, &content) != nil {
		storage.ContentUnreadable = true
		return storage
	}
	return content.ToEntity(storage)
}

func (v *vaultPersistence) createIdentity(localId string, settings entity.VaultSettings) (string, error) {
	secret, err := agecrypt.GenerateIdentity()
	if err != nil {
		return "", errs.InternalServerError(err, "VLT-17500", "the vault key could not be generated")
	}
	if settings.Mode == vaultmode.Keychain {
		if err := v.keychain.Set(vaultKeychainService, localId, secret); err != nil {
			return "", errs.InternalServerError(err, "VLT-16500", "the vault key could not be stored in the OS keychain")
		}
		return secret, nil
	}
	passphrase, unlocked := v.passphrase()
	if !unlocked {
		return "", errs.ConflictError("the vault is locked; unlock it with the passphrase", "VLT-29409")
	}
	wrapped, err := agecrypt.EncryptWithPassphrase([]byte(secret), passphrase, properties.Properties().Vault.ScryptLogN)
	if err != nil {
		return "", errs.InternalServerError(err, "VLT-18500", "the vault key could not be protected with the passphrase")
	}
	if err := writeVaultFile(vaultKeyPath(localId), wrapped); err != nil {
		return "", err
	}
	v.mu.Lock()
	v.identities[vaultKeyPath(localId)] = secret
	v.mu.Unlock()
	return secret, nil
}

func (v *vaultPersistence) passphrase() (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	passphrase, unlocked := v.passphrases[appDataPath()]
	return passphrase, unlocked
}

func (v *vaultPersistence) cachedIdentity(path string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	secret, cached := v.identities[path]
	return secret, cached
}

func writeVaultFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), appDataDirMode); err != nil {
		return errs.FileError(err, "VLT-19500")
	}
	if err := fileutils.WriteFileAtomic(path, data, appDataFileMode); err != nil {
		return errs.FileError(err, "VLT-21500")
	}
	return nil
}

func vaultPath(localId string) string {
	return appDataPath(workspacesDataDir, localId, vaultFile)
}

func vaultKeyPath(localId string) string {
	return appDataPath(workspacesDataDir, localId, vaultKeyFile)
}
