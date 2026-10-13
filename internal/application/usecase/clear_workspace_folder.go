package usecase

import "context"

type ClearWorkspaceFolder interface {
	ClearFolder(ctx context.Context, root string) error
}
