package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveCookieAllowlist interface {
	SaveCookieAllowlist(ctx context.Context, localId string, allowlist entity.CookieAllowlist) error
}
