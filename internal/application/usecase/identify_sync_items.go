package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type IdentifySyncItems interface {
	IdentifyItems(ctx context.Context, files map[string][]byte) map[string]entity.ItemRef
}
