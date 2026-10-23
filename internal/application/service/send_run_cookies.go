package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

func (r sendRun) Cookies(ctx context.Context, url string) ([]entity.Cookie, error) {
	logger.Info(ctx, "Started", map[string]any{"url": url})

	if err := r.allowCookies(ctx, url); err != nil {
		return nil, err
	}
	cookies, err := r.service.cookies.manageCookieJar.CookiesFor(ctx, r.jarId, url)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"url": url, "cookies": len(cookies)})
	return cookies, nil
}

func (r sendRun) SetCookie(ctx context.Context, url, setCookie string) (*entity.Cookie, error) {
	logger.Info(ctx, "Started", map[string]any{"url": url})

	if err := r.allowCookies(ctx, url); err != nil {
		return nil, err
	}
	cookie, err := r.service.cookies.manageCookieJar.SetCookie(ctx, r.jarId, url, setCookie)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"url": url, "result": *cookie})
	return cookie, nil
}

func (r sendRun) Unset(ctx context.Context, url, name string) error {
	logger.Info(ctx, "Started", map[string]any{"url": url, "name": name})

	cookies, err := r.Cookies(ctx, url)
	if err != nil {
		return err
	}
	r.service.cookies.manageCookieJar.RemoveCookies(ctx, r.jarId, entity.FirstCookieNamed(cookies, name))

	logger.Info(ctx, "Finished", map[string]any{"url": url, "name": name})
	return nil
}

func (r sendRun) Clear(ctx context.Context, url string) error {
	logger.Info(ctx, "Started", map[string]any{"url": url})

	cookies, err := r.Cookies(ctx, url)
	if err != nil {
		return err
	}
	r.service.cookies.manageCookieJar.RemoveCookies(ctx, r.jarId, cookies)

	logger.Info(ctx, "Finished", map[string]any{"url": url, "cookies": len(cookies)})
	return nil
}

func (r sendRun) allowCookies(ctx context.Context, url string) error {
	allowlist, err := r.service.findCookieAllowlist.FindCookieAllowlist(ctx, r.localId)
	if err != nil {
		return err
	}
	host := entity.ParseRequestUrl(url).Hostname()
	if !allowlist.Allows(host) {
		return usecase.RejectScriptCookieAccess(host)
	}
	return nil
}
