package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageCookieJar interface {
	JarState(ctx context.Context, localId string) entity.CookieJarState
	LoadCookies(ctx context.Context, localId string, cookies []entity.Cookie)
	MarkCookiesSaved(ctx context.Context, localId string, revision uint64)
	CookiesFor(ctx context.Context, localId, url string) ([]entity.Cookie, error)
	SetCookie(ctx context.Context, localId, url, setCookie string) (*entity.Cookie, error)
	RemoveCookies(ctx context.Context, localId string, cookies []entity.Cookie)
	DropJar(ctx context.Context, jarId string)
}
