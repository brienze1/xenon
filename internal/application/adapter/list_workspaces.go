package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ListWorkspacesService interface {
	List(ctx context.Context) (*entity.AppRegistry, error)
}
