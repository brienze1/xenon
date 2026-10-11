package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveGitAccount interface {
	Save(ctx context.Context, account entity.GitAccount) error
}
