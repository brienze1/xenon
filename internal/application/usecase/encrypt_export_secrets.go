package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type EncryptExportSecrets interface {
	Encrypt(ctx context.Context, secrets entity.SecretBundle, passphrase string) (string, error)
}
