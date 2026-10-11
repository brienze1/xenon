package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ListRemote interface {
	List(ctx context.Context, access entity.GitAccess) (*entity.RemoteInfo, error)
}
