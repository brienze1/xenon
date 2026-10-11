package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageCookies struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findCookieAllowlist   usecase.FindCookieAllowlist
	saveCookieAllowlist   usecase.SaveCookieAllowlist
	jar                   cookieJar
}

func NewManageCookiesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	manageCookieJar usecase.ManageCookieJar,
	findCookies usecase.FindCookies,
	saveCookies usecase.SaveCookies,
	findCookieAllowlist usecase.FindCookieAllowlist,
	saveCookieAllowlist usecase.SaveCookieAllowlist,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageCookiesService {
	return &manageCookies{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findCookieAllowlist:   findCookieAllowlist,
		saveCookieAllowlist:   saveCookieAllowlist,
		jar: cookieJar{
			manageCookieJar:       manageCookieJar,
			findCookies:           findCookies,
			saveCookies:           saveCookies,
			publishWorkspaceEvent: publishWorkspaceEvent,
			vaults: vaultWriter{
				findVault:            findVault,
				saveVault:            saveVault,
				findVaultSettings:    findVaultSettings,
				saveVaultSettings:    saveVaultSettings,
				probeKeychain:        probeKeychain,
				publishVariableEvent: publishVariableEvent,
			},
		},
	}
}

func (m *manageCookies) List(ctx context.Context, localId string) ([]entity.CookieDomain, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	var domains []entity.CookieDomain
	err := m.change(ctx, localId, func() error {
		domains = entity.GroupCookiesByDomain(m.jar.manageCookieJar.JarState(ctx, localId).Cookies)
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "domains": len(domains)})
	return domains, nil
}

func (m *manageCookies) Upsert(ctx context.Context, localId, domain, setCookie string, replaces *entity.Cookie) (*entity.Cookie, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "domain": domain})

	var stored *entity.Cookie
	err := m.change(ctx, localId, func() error {
		cookie, err := m.jar.manageCookieJar.SetCookie(ctx, localId, usecase.CookieDomainUrl(domain), setCookie)
		if err != nil {
			return err
		}
		if replaces != nil && !cookie.Is(replaces.Domain, replaces.Path, replaces.Name) {
			m.jar.manageCookieJar.RemoveCookies(ctx, localId, []entity.Cookie{*replaces})
		}
		stored = cookie
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "domain": domain, "result": *stored})
	return stored, nil
}

func (m *manageCookies) Delete(ctx context.Context, localId string, cookie entity.Cookie) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "cookie": cookie})

	err := m.change(ctx, localId, func() error {
		m.jar.manageCookieJar.RemoveCookies(ctx, localId, []entity.Cookie{cookie})
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "cookie": cookie})
	return nil
}

func (m *manageCookies) DeleteDomain(ctx context.Context, localId, domain string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "domain": domain})

	err := m.change(ctx, localId, func() error {
		cookies := m.jar.manageCookieJar.JarState(ctx, localId).Cookies
		m.jar.manageCookieJar.RemoveCookies(ctx, localId, entity.CookiesOn(cookies, domain))
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "domain": domain})
	return nil
}

func (m *manageCookies) Clear(ctx context.Context, localId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	err := m.change(ctx, localId, func() error {
		m.jar.manageCookieJar.RemoveCookies(ctx, localId, m.jar.manageCookieJar.JarState(ctx, localId).Cookies)
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (m *manageCookies) GetAllowlist(ctx context.Context, localId string) (*entity.CookieAllowlist, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	allowlist, err := m.findCookieAllowlist.FindCookieAllowlist(ctx, localId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": *allowlist})
	return allowlist, nil
}

func (m *manageCookies) SaveAllowlist(ctx context.Context, localId string, domains []string) (*entity.CookieAllowlist, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "domains": domains})

	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	allowlist := entity.NewCookieAllowlist(domains)
	if err := m.saveCookieAllowlist.SaveCookieAllowlist(ctx, localId, allowlist); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": allowlist})
	return &allowlist, nil
}

func (m *manageCookies) change(ctx context.Context, localId string, apply func() error) error {
	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := m.jar.load(ctx, localId); err != nil {
		return err
	}
	before := m.jar.revision(ctx, localId)
	if err := apply(); err != nil {
		return err
	}
	return m.jar.settle(ctx, localId, before)
}
