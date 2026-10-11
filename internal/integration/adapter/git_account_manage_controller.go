package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type GitAccountManageController interface {
	List(ctx context.Context) ([]dto.GitAccount, error)
	Update(ctx context.Context, request dto.UpdateGitAccount) (*dto.GitAccount, error)
	Remove(ctx context.Context, request dto.GitAccountReference) error
}
