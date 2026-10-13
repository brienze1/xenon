package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type importSecrets struct {
	decryptExportSecrets usecase.DecryptExportSecrets
}

func (i importSecrets) unlocked(ctx context.Context, source entity.ImportSource, plan entity.ImportPlan) (*entity.ImportSource, error) {
	if !plan.NeedsPassphrase(source) {
		return &source, nil
	}
	if !plan.HasPassphrase() {
		return nil, usecase.RejectMissingImportPassphrase()
	}
	secrets, err := i.decryptExportSecrets.Decrypt(ctx, source.EncryptedSecrets, plan.Passphrase)
	if err != nil {
		return nil, err
	}
	unlocked := source.WithSecrets(*secrets)
	return &unlocked, nil
}
