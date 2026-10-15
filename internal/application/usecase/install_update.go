package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type InstallUpdate interface {
	InstallUpdate(ctx context.Context, installation entity.Installation, packagePath string) error
}
