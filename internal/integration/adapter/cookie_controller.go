package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type CookieController interface {
	List(ctx context.Context, request dto.CookieJarReference) (dto.CookieDomainList, error)
	Upsert(ctx context.Context, request dto.UpsertCookie) (*dto.Cookie, error)
	Delete(ctx context.Context, request dto.DeleteCookie) error
	DeleteDomain(ctx context.Context, request dto.DeleteCookieDomain) error
	Clear(ctx context.Context, request dto.CookieJarReference) error
	GetAllowlist(ctx context.Context, request dto.CookieJarReference) (*dto.CookieAllowlist, error)
	SaveAllowlist(ctx context.Context, request dto.SaveCookieAllowlist) (*dto.CookieAllowlist, error)
}
