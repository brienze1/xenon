package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishWorkspaceEvent interface {
	TreeChanged(ctx context.Context, localId string) error
	RequestChangedOnDisk(ctx context.Context, localId string, change entity.RequestChange) error
	FolderMissing(ctx context.Context, localId string) error
	CookiesChanged(ctx context.Context, localId string) error
}
