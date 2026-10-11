package usecase

import "context"

type DeleteGitCredential interface {
	DeleteCredential(ctx context.Context, accountId string) error
}
