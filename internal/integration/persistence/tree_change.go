package persistence

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/logger"
)

func (t *treePersistence) Inspect(ctx context.Context, root, path string) (*entity.FileChange, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "path": path})
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	if data, err := os.ReadFile(path); err == nil && t.index.IsOwnWrite(path, data) {
		logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "ownWrite": true})
		return &entity.FileChange{OwnWrite: true}, nil
	}
	affected := t.requestsAt(root, path)
	index, err := t.index.refresh(root)
	if err != nil {
		return nil, err
	}
	change := &entity.FileChange{Requests: make([]entity.RequestChange, 0, len(affected))}
	for _, id := range affected {
		change.Requests = append(change.Requests, requestChange(id, index.items[id]))
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "requests": len(change.Requests)})
	return change, nil
}

func (t *treePersistence) requestsAt(root, path string) []string {
	ids := make([]string, 0)
	index, found := t.index.roots[root]
	if !found || filepath.Clean(path) == filepath.Clean(root) {
		return ids
	}
	for id, item := range index.items {
		if item.kind == itemkind.Request && isSameOrInside(item.path, path) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func requestChange(id string, item indexedItem) entity.RequestChange {
	data, err := os.ReadFile(item.path)
	if err != nil {
		return entity.RequestChange{RequestID: id, Deleted: true}
	}
	return entity.RequestChange{RequestID: id, Revision: revisionOf(data)}
}

func isSameOrInside(child, parent string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
