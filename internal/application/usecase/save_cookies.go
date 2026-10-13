package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveCookies interface {
	SaveCookies(ctx context.Context, localId string, cookies []entity.Cookie) error
}
