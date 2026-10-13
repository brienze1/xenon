package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type CookieService struct {
	cookie adapter.CookieController
}

func NewCookieService(cookie adapter.CookieController) *CookieService {
	return &CookieService{cookie: cookie}
}

func (c *CookieService) List(ctx context.Context, request dto.CookieJarReference) (dto.CookieDomainList, error) {
	return c.cookie.List(ctx, request)
}

func (c *CookieService) Upsert(ctx context.Context, request dto.UpsertCookie) (*dto.Cookie, error) {
	return c.cookie.Upsert(ctx, request)
}

func (c *CookieService) Delete(ctx context.Context, request dto.DeleteCookie) error {
	return c.cookie.Delete(ctx, request)
}

func (c *CookieService) DeleteDomain(ctx context.Context, request dto.DeleteCookieDomain) error {
	return c.cookie.DeleteDomain(ctx, request)
}

func (c *CookieService) Clear(ctx context.Context, request dto.CookieJarReference) error {
	return c.cookie.Clear(ctx, request)
}

func (c *CookieService) GetAllowlist(ctx context.Context, request dto.CookieJarReference) (*dto.CookieAllowlist, error) {
	return c.cookie.GetAllowlist(ctx, request)
}

func (c *CookieService) SaveAllowlist(ctx context.Context, request dto.SaveCookieAllowlist) (*dto.CookieAllowlist, error) {
	return c.cookie.SaveAllowlist(ctx, request)
}
