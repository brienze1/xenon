package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type WorkspaceIndexCache interface {
	Lock(ctx context.Context)
	Unlock(ctx context.Context)
	// IsOwnWrite reports whether path was written by Xenon with exactly data in the last two seconds.
	IsOwnWrite(ctx context.Context, path string, data []byte) bool
	RecordWrite(ctx context.Context, path string, data []byte)
	// IsOwnMove reports whether path is, or is inside, a source or target Xenon moved in the last two seconds.
	IsOwnMove(ctx context.Context, path string) bool
	RecordMove(ctx context.Context, source, target string)
	Cached(ctx context.Context, root string) (*entity.TreeIndex, bool)
	Current(ctx context.Context, root string) (*entity.TreeIndex, error)
	Refresh(ctx context.Context, root string) (*entity.TreeIndex, error)
	Resolve(ctx context.Context, root, id string) (*entity.TreeIndex, entity.IndexedItem, bool, error)
}
