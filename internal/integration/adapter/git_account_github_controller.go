package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type GitAccountGitHubController interface {
	StartGitHubSignIn(ctx context.Context) (*dto.GitHubDeviceCode, error)
	PollGitHubSignIn(ctx context.Context, request dto.GitHubSignInSession) (*dto.GitHubSignInResult, error)
	UploadPublicKey(ctx context.Context, request dto.UploadGitPublicKey) error
}
