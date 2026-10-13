package usecase

import "context"

type MoveBranch interface {
	MoveBranch(ctx context.Context, clonePath, branch, hash string) error
}
