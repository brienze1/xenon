package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageEnvironmentService interface {
	List(ctx context.Context, localId string) (*entity.EnvironmentList, error)
	SetActive(ctx context.Context, localId, id string) (string, error)
}
