package persistence

import (
	"context"
	"encoding/json"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const cookiesFile = "cookies.age"

func (v *vaultPersistence) FindCookies(ctx context.Context, localId string) (*entity.StoredCookies, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	identity, err := v.cookieIdentity(localId)
	if err != nil {
		return nil, err
	}
	stored := &entity.StoredCookies{Cookies: []entity.Cookie{}, Available: !identity.sealed}
	if identity.found && !identity.sealed {
		stored.Cookies = readCookies(cookiesPath(localId), identity.secret)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "available": stored.Available, "cookies": len(stored.Cookies)})
	return stored, nil
}

func (v *vaultPersistence) SaveCookies(ctx context.Context, localId string, cookies []entity.Cookie) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "cookies": len(cookies)})

	identity, err := v.cookieIdentity(localId)
	if err != nil {
		return err
	}
	if !identity.found || identity.sealed || identity.unreadable {
		return errs.ConflictError("the vault is locked or has no key; the cookies stay in memory", "CKE-08409")
	}
	content, _ := json.Marshal(model.NewCookies(cookies))
	encrypted, err := agecrypt.EncryptForIdentity(content, identity.secret)
	if err != nil {
		return errs.InternalServerError(err, "CKE-13500", "the cookies could not be encrypted")
	}
	if err := writeVaultFile(cookiesPath(localId), encrypted); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "cookies": len(cookies)})
	return nil
}

func (v *vaultPersistence) cookieIdentity(localId string) (*vaultIdentity, error) {
	settings, err := readVaultSettings()
	if err != nil {
		return nil, err
	}
	return v.identity(localId, *settings)
}

func readCookies(path, secret string) []entity.Cookie {
	cookies := make([]entity.Cookie, 0)
	data, err := os.ReadFile(path)
	if err != nil {
		return cookies
	}
	plaintext, err := agecrypt.DecryptWithIdentity(data, secret)
	var stored []model.Cookie
	if err != nil || json.Unmarshal(plaintext, &stored) != nil {
		return cookies
	}
	for _, cookie := range stored {
		cookies = append(cookies, cookie.ToEntity())
	}
	return cookies
}

func cookiesPath(localId string) string {
	return appDataPath(workspacesDataDir, localId, cookiesFile)
}
