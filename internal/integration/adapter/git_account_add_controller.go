package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type GitAccountAddController interface {
	AddToken(ctx context.Context, request dto.AddGitToken) (*dto.GitAccount, error)
	AddSSH(ctx context.Context, request dto.AddGitSsh) (*dto.GitSshAccount, error)
}
