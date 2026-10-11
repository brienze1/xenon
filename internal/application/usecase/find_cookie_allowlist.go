package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindCookieAllowlist interface {
	FindCookieAllowlist(ctx context.Context, localId string) (*entity.CookieAllowlist, error)
}
