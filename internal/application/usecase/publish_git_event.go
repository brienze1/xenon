package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishGitEvent interface {
	HostKeyUnknown(ctx context.Context, repository string, key entity.HostKey) error
}
