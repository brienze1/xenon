package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveTreeItem interface {
	SaveContainer(ctx context.Context, root string, container entity.Collection) (*entity.Collection, error)
}
