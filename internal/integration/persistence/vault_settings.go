package persistence

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/encryption"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	vaultSettingsKey     = "vault"
	passphraseCheckToken = "xenon-vault-passphrase-check"
)

func (v *vaultPersistence) FindSettings(ctx context.Context) (*entity.VaultSettings, error) {
	logger.Debug(ctx, "Started")

	settings, err := readVaultSettings()
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", *settings)
	return settings, nil
}

func (v *vaultPersistence) SaveSettings(ctx context.Context, settings entity.VaultSettings) error {
	logger.Debug(ctx, "Started", settings)

	stored, err := readAppJson()
	if err != nil {
		return err
	}
	stored[vaultSettingsKey], _ = json.Marshal(model.NewVaultSettings(settings))
	if err := os.MkdirAll(appDataPath(), appDataDirMode); err != nil {
		return errs.FileError(err, "VLT-26500")
	}
	data, _ := json.MarshalIndent(stored, "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(registryFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "VLT-27500")
	}

	logger.Debug(ctx, "Finished", settings)
	return nil
}

func (v *vaultPersistence) Seal(ctx context.Context, passphrase string) (string, error) {
	logger.Debug(ctx, "Started")

	sealed, err := encryption.EncryptWithPassphrase([]byte(passphraseCheckToken), passphrase, properties.Properties().Vault.ScryptLogN)
	if err != nil {
		return "", errs.InternalServerError(err, "VLT-28500", "the passphrase check could not be created")
	}

	logger.Debug(ctx, "Finished")
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (v *vaultPersistence) Verify(ctx context.Context, check, passphrase string) bool {
	logger.Debug(ctx, "Started")

	sealed, err := base64.StdEncoding.DecodeString(check)
	if err != nil {
		return false
	}
	token, err := encryption.DecryptWithPassphrase(sealed, passphrase)
	verified := err == nil && string(token) == passphraseCheckToken

	logger.Debug(ctx, "Finished", map[string]any{"verified": verified})
	return verified
}

func readVaultSettings() (*entity.VaultSettings, error) {
	stored, err := readAppJson()
	if err != nil {
		return nil, err
	}
	var settings model.VaultSettings
	if raw, found := stored[vaultSettingsKey]; found {
		_ = json.Unmarshal(raw, &settings)
	}
	parsed := settings.ToEntity()
	return &parsed, nil
}

func readAppJson() (map[string]json.RawMessage, error) {
	stored := map[string]json.RawMessage{}
	data, err := os.ReadFile(appDataPath(registryFile))
	if errors.Is(err, fs.ErrNotExist) {
		return stored, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "VLT-25500")
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.InternalServerError(err, "VLT-30500", "app.json is not valid JSON")
	}
	return stored, nil
}
