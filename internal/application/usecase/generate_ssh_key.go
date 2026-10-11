package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type GenerateSshKey interface {
	Generate(ctx context.Context, accountId string) (*entity.SshKey, error)
}
