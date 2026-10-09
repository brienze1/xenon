package adapter

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/persistence/model"
)

type WorkspaceIndex interface {
	sync.Locker
	// IsOwnWrite reports whether path was written by Xenon with exactly data in the last two seconds.
	IsOwnWrite(path string, data []byte) bool
	RecordWrite(path string, data []byte)
	Cached(root string) (*model.TreeIndex, bool)
	Current(root string) (*model.TreeIndex, error)
	Refresh(root string) (*model.TreeIndex, error)
	Resolve(root, id string) (*model.TreeIndex, model.IndexedItem, bool, error)
}
