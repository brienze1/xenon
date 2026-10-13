package usecase

import "context"

type DeleteTreeItem interface {
	Delete(ctx context.Context, root, id string) error
}
