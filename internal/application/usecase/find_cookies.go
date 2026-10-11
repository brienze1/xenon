package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindCookies interface {
	FindCookies(ctx context.Context, localId string) (*entity.StoredCookies, error)
}
