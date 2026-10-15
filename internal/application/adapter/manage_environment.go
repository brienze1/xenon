package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageEnvironmentService interface {
	List(ctx context.Context, localId string) (*entity.EnvironmentList, error)
	Get(ctx context.Context, localId, id string) (*entity.EnvironmentDetails, error)
	Create(ctx context.Context, localId, name string) (*entity.EnvironmentDetails, error)
	Update(ctx context.Context, localId string, settings entity.Environment) (*entity.EnvironmentDetails, error)
	Duplicate(ctx context.Context, localId, id string) (*entity.EnvironmentDetails, error)
	Delete(ctx context.Context, localId, id string) (string, error)
}
