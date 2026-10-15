package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishUpdateStatus interface {
	PublishUpdateStatus(ctx context.Context, status entity.UpdateStatus)
}
