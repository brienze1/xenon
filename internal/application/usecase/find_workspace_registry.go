package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindWorkspaceRegistry interface {
	Find(ctx context.Context) (*entity.AppRegistry, error)
	FindWorkspace(ctx context.Context, localId string) (*entity.RegisteredWorkspace, error)
}
