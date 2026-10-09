package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const ownWriteWindow = 2 * time.Second

type WorkspaceIndex struct {
	roots    map[string]*rootIndex
	writes   map[string]ownWrite
	mu       sync.Mutex
	writesMu sync.Mutex
}

type rootIndex struct {
	items        map[string]indexedItem
	reidentified map[string]string
	tree         entity.Tree
}

type indexedItem struct {
	path string
	kind enums.ItemKind
}

type ownWrite struct {
	at   time.Time
	hash string
}

func NewWorkspaceIndex() *WorkspaceIndex {
	return &WorkspaceIndex{roots: map[string]*rootIndex{}, writes: map[string]ownWrite{}}
}

// IsOwnWrite reports whether path was written by Xenon with exactly data in the last two seconds.
func (w *WorkspaceIndex) IsOwnWrite(path string, data []byte) bool {
	w.writesMu.Lock()
	defer w.writesMu.Unlock()
	write, found := w.writes[path]
	return found && write.hash == revisionOf(data) && timeutils.Now().Sub(write.at) <= ownWriteWindow
}

func (w *WorkspaceIndex) recordWrite(path string, data []byte) {
	w.writesMu.Lock()
	defer w.writesMu.Unlock()
	w.writes[path] = ownWrite{at: timeutils.Now(), hash: revisionOf(data)}
}

func (w *WorkspaceIndex) current(root string) (*rootIndex, error) {
	if index, found := w.roots[root]; found {
		return index, nil
	}
	return w.refresh(root)
}

func (w *WorkspaceIndex) refresh(root string) (*rootIndex, error) {
	previous := map[string]string{}
	if index, found := w.roots[root]; found {
		previous = index.reidentified
	}
	index, err := scanTree(root, previous)
	if err != nil {
		return nil, err
	}
	w.roots[root] = index
	return index, nil
}

func (w *WorkspaceIndex) resolve(root, id string) (*rootIndex, indexedItem, bool, error) {
	if index, found := w.roots[root]; found {
		if item, known := index.items[id]; known && exists(item.path) {
			return index, item, true, nil
		}
	}
	index, err := w.refresh(root)
	if err != nil {
		return nil, indexedItem{}, false, err
	}
	item, known := index.items[id]
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
