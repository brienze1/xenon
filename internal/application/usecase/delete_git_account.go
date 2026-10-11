package usecase

import "context"

type DeleteGitAccount interface {
	Delete(ctx context.Context, id string) error
}
