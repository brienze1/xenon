package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type WritePostmanExport interface {
	WriteCollection(ctx context.Context, root entity.ExportNode) ([]byte, error)
	WriteEnvironment(ctx context.Context, environment entity.Environment) ([]byte, error)
	WriteGlobals(ctx context.Context, workspace entity.Workspace) ([]byte, error)
}
