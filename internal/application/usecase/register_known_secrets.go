package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RegisterKnownSecrets interface {
	Register(ctx context.Context, values ...string) entity.KnownSecrets
}
