package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type VerifyGitToken interface {
	Verify(ctx context.Context, access entity.GitAccess) (*entity.RemoteInfo, error)
}
