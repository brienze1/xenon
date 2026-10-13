package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadSshKey interface {
	Read(ctx context.Context, path, passphrase string) (*entity.SshKey, error)
}
