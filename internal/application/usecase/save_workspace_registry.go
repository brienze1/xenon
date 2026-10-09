package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveWorkspaceRegistry interface {
	Save(ctx context.Context, registry entity.AppRegistry) error
}
