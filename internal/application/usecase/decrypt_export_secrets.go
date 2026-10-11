package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type DecryptExportSecrets interface {
	Decrypt(ctx context.Context, armored, passphrase string) (*entity.SecretBundle, error)
}
