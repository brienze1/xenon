package usecase

import "context"

type DeleteEnvironment interface {
	Delete(ctx context.Context, root, id string) error
}
