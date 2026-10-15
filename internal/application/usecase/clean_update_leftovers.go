package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CleanUpdateLeftovers interface {
	CleanUpdateLeftovers(ctx context.Context, installation entity.Installation)
}
