package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindGitHubUser interface {
	FindUser(ctx context.Context, token string) (*entity.GitHubUser, error)
}
