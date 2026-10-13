package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindGitAccounts interface {
	FindAll(ctx context.Context) ([]entity.GitAccount, error)
	Find(ctx context.Context, id string) (*entity.GitAccount, error)
}
