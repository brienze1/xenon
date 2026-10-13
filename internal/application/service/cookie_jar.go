package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type cookieJar struct {
	manageCookieJar       usecase.ManageCookieJar
	findCookies           usecase.FindCookies
	saveCookies           usecase.SaveCookies
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
	vaults                vaultWriter
}

func (c cookieJar) load(ctx context.Context, localId string) error {
	if c.manageCookieJar.JarState(ctx, localId).Loaded {
		return nil
	}
	stored, err := c.findCookies.FindCookies(ctx, localId)
	if err != nil {
		return err
	}
	if stored.Available {
		c.manageCookieJar.LoadCookies(ctx, localId, stored.Cookies)
	}
	return nil
}

func (c cookieJar) revision(ctx context.Context, localId string) uint64 {
	return c.manageCookieJar.JarState(ctx, localId).Revision
}

func (c cookieJar) header(ctx context.Context, localId string, request entity.EffectiveRequest) string {
	if !request.UsesCookieJar() {
		return ""
	}
	cookies, err := c.manageCookieJar.CookiesFor(ctx, localId, request.URL)
	if err != nil {
		return ""
	}
	return entity.CookieHeader(cookies)
}

func (c cookieJar) settle(ctx context.Context, localId string, before uint64) error {
	if err := c.load(ctx, localId); err != nil {
		return err
	}
	state := c.manageCookieJar.JarState(ctx, localId)
	if state.Loaded && state.HasUnsavedChanges() {
		c.save(ctx, localId, state)
	}
	if !state.ChangedSince(before) {
		return nil
	}
	return c.publishWorkspaceEvent.CookiesChanged(ctx, localId)
}

func (c cookieJar) save(ctx context.Context, localId string, state entity.CookieJarState) {
	if err := c.write(ctx, localId, entity.PersistentCookies(state.Cookies)); err != nil {
		logger.Warn(ctx, err, "The cookies stay in memory until the vault can be written", map[string]any{"localId": localId})
		return
	}
	c.manageCookieJar.MarkCookiesSaved(ctx, localId, state.Revision)
}

func (c cookieJar) write(ctx context.Context, localId string, cookies []entity.Cookie) error {
	vault, err := c.vaults.findVault.Find(ctx, localId)
	if err != nil {
		return err
	}
	if vault.IsEmpty() {
		if len(cookies) == 0 {
			return nil
		}
		if _, err := c.vaults.update(ctx, localId, keepVault); err != nil {
			return err
		}
	}
	return c.saveCookies.SaveCookies(ctx, localId, cookies)
}

func keepVault(vault entity.Vault) entity.Vault {
	return vault
}
