package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type TrustHostKey interface {
	Trust(ctx context.Context, key entity.HostKey) error
}
