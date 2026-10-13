package usecase

import "context"

type SaveGitCredential interface {
	CheckCredentialStore(ctx context.Context) error
	SaveCredential(ctx context.Context, accountId, secret string) error
}
