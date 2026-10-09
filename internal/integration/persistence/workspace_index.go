package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceIndexPersistence struct {
	roots    map[string]*entity.TreeIndex
	writes   map[string]entity.RecordedWrite
	mu       sync.Mutex
	writesMu sync.Mutex
}

func NewWorkspaceIndexPersistence() adapter.WorkspaceIndexPersistence {
	return &workspaceIndexPersistence{roots: map[string]*entity.TreeIndex{}, writes: map[string]entity.RecordedWrite{}}
}

func (w *workspaceIndexPersistence) Lock(ctx context.Context) {
	logger.Debug(ctx, "Started")

	w.mu.Lock()

	logger.Debug(ctx, "Finished")
}

func (w *workspaceIndexPersistence) Unlock(ctx context.Context) {
	logger.Debug(ctx, "Started")

	w.mu.Unlock()

	logger.Debug(ctx, "Finished")
}

func (w *workspaceIndexPersistence) IsOwnWrite(ctx context.Context, path string, data []byte) bool {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "size": len(data)})

	w.writesMu.Lock()
	write, found := w.writes[path]
	w.writesMu.Unlock()
	ownWrite := found && write.Echoes(revisionOf(data))

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "ownWrite": ownWrite})
	return ownWrite
}

func (w *workspaceIndexPersistence) RecordWrite(ctx context.Context, path string, data []byte) {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "size": len(data)})

	w.writesMu.Lock()
	w.writes[path] = entity.NewRecordedWrite(revisionOf(data))
	w.writesMu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"path": path})
}

func (w *workspaceIndexPersistence) Cached(ctx context.Context, root string) (*entity.TreeIndex, bool) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	index, found := w.roots[root]

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "found": found})
	return index, found
}

func (w *workspaceIndexPersistence) Current(ctx context.Context, root string) (*entity.TreeIndex, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	if index, found := w.roots[root]; found {
		logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": index.Size()})
		return index, nil
	}
	index, err := w.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": index.Size()})
	return index, nil
}

func (w *workspaceIndexPersistence) Refresh(ctx context.Context, root string) (*entity.TreeIndex, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	previous := map[string]string{}
	if index, found := w.roots[root]; found {
		previous = index.Reidentified
	}
	index, err := scanTree(root, previous)
	if err != nil {
		return nil, err
	}
	w.roots[root] = index

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": index.Size()})
	return index, nil
}

func (w *workspaceIndexPersistence) Resolve(ctx context.Context, root, id string) (*entity.TreeIndex, entity.IndexedItem, bool, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})

	if index, found := w.roots[root]; found {
		if item, known := index.Item(id); known && exists(item.Path) {
			logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id, "found": true})
			return index, item, true, nil
		}
	}
	index, err := w.Refresh(ctx, root)
	if err != nil {
		return nil, entity.IndexedItem{}, false, err
	}
	item, known := index.Item(id)

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id, "found": known})
	return index, item, known, nil
}

func revisionOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
