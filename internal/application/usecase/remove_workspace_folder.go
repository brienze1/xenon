package usecase

import "context"

type RemoveWorkspaceFolder interface {
	RemoveFolder(ctx context.Context, path string) error
}
