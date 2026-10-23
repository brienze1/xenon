package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ClaimActiveRun interface {
	Claim(ctx context.Context, run entity.Run, control RunControl) error
}
