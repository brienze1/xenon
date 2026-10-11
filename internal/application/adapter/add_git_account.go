package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type AddGitAccountService interface {
	AddToken(ctx context.Context, draft entity.GitAccount, token, repository string) (*entity.GitAccount, error)
	AddSsh(ctx context.Context, draft entity.GitAccount, keyPath, passphrase string) (*entity.GitAccount, string, error)
}
