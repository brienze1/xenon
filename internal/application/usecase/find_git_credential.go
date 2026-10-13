package usecase

import "context"

type FindGitCredential interface {
	FindCredential(ctx context.Context, accountId string) (string, error)
}
