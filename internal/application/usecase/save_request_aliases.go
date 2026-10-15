package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRequestAliases interface {
	SaveAliases(ctx context.Context, root string, request entity.Request) (*entity.Request, error)
}
