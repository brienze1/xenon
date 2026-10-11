package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

// ScriptCookieJar is pm.cookies.jar(): the workspace jar, limited to the allowlisted domains.
type ScriptCookieJar interface {
	Cookies(ctx context.Context, url string) ([]entity.Cookie, error)
	SetCookie(ctx context.Context, url, setCookie string) (*entity.Cookie, error)
	Unset(ctx context.Context, url, name string) error
	Clear(ctx context.Context, url string) error
}
