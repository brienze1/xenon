package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const ownWriteWindow = 2 * time.Second

type workspaceIndex struct {
	roots    map[string]*model.TreeIndex
	writes   map[string]ownWrite
	mu       sync.Mutex
	writesMu sync.Mutex
}

type ownWrite struct {
	at   time.Time
	hash string
}

func NewWorkspaceIndex() adapter.WorkspaceIndex {
	return &workspaceIndex{roots: map[string]*model.TreeIndex{}, writes: map[string]ownWrite{}}
}

func (w *workspaceIndex) Lock() {
	w.mu.Lock()
}

func (w *workspaceIndex) Unlock() {
	w.mu.Unlock()
}

func (w *workspaceIndex) IsOwnWrite(path string, data []byte) bool {
	w.writesMu.Lock()
	defer w.writesMu.Unlock()
	write, found := w.writes[path]
	return found && write.hash == revisionOf(data) && timeutils.Now().Sub(write.at) <= ownWriteWindow
}

func (w *workspaceIndex) RecordWrite(path string, data []byte) {
	w.writesMu.Lock()
	defer w.writesMu.Unlock()
	w.writes[path] = ownWrite{at: timeutils.Now(), hash: revisionOf(data)}
}

func (w *workspaceIndex) Cached(root string) (*model.TreeIndex, bool) {
	index, found := w.roots[root]
	return index, found
}

func (w *workspaceIndex) Current(root string) (*model.TreeIndex, error) {
	if index, found := w.roots[root]; found {
		return index, nil
	}
	return w.Refresh(root)
}

func (w *workspaceIndex) Refresh(root string) (*model.TreeIndex, error) {
	previous := map[string]string{}
	if index, found := w.roots[root]; found {
		previous = index.Reidentified
	}
	index, err := scanTree(root, previous)
	if err != nil {
		return nil, err
	}
	w.roots[root] = index
	return index, nil
}

func (w *workspaceIndex) Resolve(root, id string) (*model.TreeIndex, model.IndexedItem, bool, error) {
	if index, found := w.roots[root]; found {
		if item, known := index.Items[id]; known && exists(item.Path) {
			return index, item, true, nil
		}
	}
	index, err := w.Refresh(root)
	if err != nil {
		return nil, model.IndexedItem{}, false, err
	}
	item, known := index.Items[id]
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
