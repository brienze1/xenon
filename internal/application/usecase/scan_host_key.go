package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ScanHostKey interface {
	Scan(ctx context.Context, repository string) (*entity.HostKey, error)
}
