package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type VerifyGitHostService interface {
	TestConnection(ctx context.Context, accountId, repository string) (*entity.RemoteInfo, error)
	ConfirmHostKey(ctx context.Context, repository, fingerprint string, accept bool) error
}
