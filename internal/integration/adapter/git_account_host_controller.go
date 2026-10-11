package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type GitAccountHostController interface {
	TestConnection(ctx context.Context, request dto.TestGitConnection) (*dto.GitRemoteInfo, error)
	ConfirmHostKey(ctx context.Context, request dto.ConfirmGitHostKey) error
}
