package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"slices"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceIndexCache struct {
	roots    map[string]*entity.TreeIndex
	writes   map[string]entity.RecordedWrite
	moves    []entity.RecordedMove
	mu       sync.Mutex
	writesMu sync.Mutex
}

func NewWorkspaceIndexCache() adapter.WorkspaceIndexCache {
	return &workspaceIndexCache{roots: map[string]*entity.TreeIndex{}, writes: map[string]entity.RecordedWrite{}}
}

func (w *workspaceIndexCache) Lock(ctx context.Context) {
	logger.Debug(ctx, "Started")

	w.mu.Lock()

	logger.Debug(ctx, "Finished")
}

func (w *workspaceIndexCache) Unlock(ctx context.Context) {
	logger.Debug(ctx, "Started")

	w.mu.Unlock()

	logger.Debug(ctx, "Finished")
}

func (w *workspaceIndexCache) IsOwnWrite(ctx context.Context, path string, data []byte) bool {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "size": len(data)})

	w.writesMu.Lock()
	write, found := w.writes[path]
	w.writesMu.Unlock()
	ownWrite := found && write.Echoes(revisionOf(data))

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "ownWrite": ownWrite})
	return ownWrite
}

func (w *workspaceIndexCache) RecordWrite(ctx context.Context, path string, data []byte) {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "size": len(data)})

	w.writesMu.Lock()
	w.writes[path] = entity.NewRecordedWrite(revisionOf(data))
	w.writesMu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"path": path})
}

func (w *workspaceIndexCache) IsOwnMove(ctx context.Context, path string) bool {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	w.writesMu.Lock()
	ownMove := slices.ContainsFunc(w.moves, func(move entity.RecordedMove) bool { return move.Covers(path) })
	w.writesMu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "ownMove": ownMove})
	return ownMove
}

func (w *workspaceIndexCache) RecordMove(ctx context.Context, source, target string) {
	logger.Debug(ctx, "Started", map[string]any{"source": source, "target": target})

	w.writesMu.Lock()
	w.moves = append(slices.DeleteFunc(w.moves, func(move entity.RecordedMove) bool { return !move.IsRecent() }), entity.NewRecordedMove(source, target))
	w.writesMu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"source": source, "target": target})
}

func (w *workspaceIndexCache) Cached(ctx context.Context, root string) (*entity.TreeIndex, bool) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	index, found := w.roots[root]

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "found": found})
	return index, found
}

func (w *workspaceIndexCache) Current(ctx context.Context, root string) (*entity.TreeIndex, error) {
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

func (w *workspaceIndexCache) Refresh(ctx context.Context, root string) (*entity.TreeIndex, error) {
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

func (w *workspaceIndexCache) Indexed(ctx context.Context, root string) (*entity.TreeIndex, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	index, found := w.roots[root]
	if !found {
		refreshed, err := w.Refresh(ctx, root)
		if err != nil {
			return nil, err
		}
		index = refreshed
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": index.Size()})
	return index, nil
}

func (w *workspaceIndexCache) Resolve(ctx context.Context, root, id string) (*entity.TreeIndex, entity.IndexedItem, bool, error) {
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
