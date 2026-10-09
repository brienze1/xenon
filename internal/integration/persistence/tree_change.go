package persistence

import (
	"context"
	"os"
	"path/filepath"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

func (t *treePersistence) Inspect(ctx context.Context, root, path string) (*entity.FileChange, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "path": path})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	if data, err := os.ReadFile(path); err == nil && t.index.IsOwnWrite(ctx, path, data) {
		logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "ownWrite": true})
		return &entity.FileChange{OwnWrite: true}, nil
	}
	affected := t.requestsAt(ctx, root, path)
	index, err := t.index.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}
	change := &entity.FileChange{Requests: make([]entity.RequestChange, 0, len(affected))}
	for _, id := range affected {
		item, _ := index.Item(id)
		change.Requests = append(change.Requests, requestChange(id, item))
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "requests": len(change.Requests)})
	return change, nil
}

func (t *treePersistence) requestsAt(ctx context.Context, root, path string) []string {
	index, found := t.index.Cached(ctx, root)
	if !found || filepath.Clean(path) == filepath.Clean(root) {
		return []string{}
	}
	return index.RequestIDsInside(path)
}

func requestChange(id string, item entity.IndexedItem) entity.RequestChange {
	data, err := os.ReadFile(item.Path)
	if err != nil {
		return entity.RequestChange{RequestID: id, Deleted: true}
	}
	return entity.RequestChange{RequestID: id, Revision: revisionOf(data)}
}
