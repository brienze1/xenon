package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SignInWithGitHub interface {
	Start(ctx context.Context) (*entity.GitHubDeviceCode, error)
	Poll(ctx context.Context, sessionId string) (*entity.GitHubSignIn, error)
}
