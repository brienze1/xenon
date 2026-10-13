package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SignInGitHubService interface {
	Start(ctx context.Context) (*entity.GitHubDeviceCode, error)
	Poll(ctx context.Context, sessionId string) (*entity.GitHubSignIn, *entity.GitAccount, error)
	UploadPublicKey(ctx context.Context, sshAccountId, gitHubAccountId, title string) error
}
