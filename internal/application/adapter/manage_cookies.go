package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageCookiesService interface {
	List(ctx context.Context, localId string) ([]entity.CookieDomain, error)
	Upsert(ctx context.Context, localId, domain, setCookie string, replaces *entity.Cookie) (*entity.Cookie, error)
	Delete(ctx context.Context, localId string, cookie entity.Cookie) error
	DeleteDomain(ctx context.Context, localId, domain string) error
	Clear(ctx context.Context, localId string) error
	GetAllowlist(ctx context.Context, localId string) (*entity.CookieAllowlist, error)
	SaveAllowlist(ctx context.Context, localId string, domains []string) (*entity.CookieAllowlist, error)
}
