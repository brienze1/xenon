package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReorderEnvironmentService interface {
	Reorder(ctx context.Context, localId, id string, index int) (*entity.EnvironmentList, error)
}
