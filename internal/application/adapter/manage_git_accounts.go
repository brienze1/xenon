package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageGitAccountsService interface {
	List(ctx context.Context) ([]entity.GitAccount, error)
	Update(ctx context.Context, id, displayName, commitName, commitEmail string) (*entity.GitAccount, error)
	Remove(ctx context.Context, id string) error
}
