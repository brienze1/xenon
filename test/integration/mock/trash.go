package mock

import (
	"errors"
	"sync"
)

type Trash struct {
	failure error
	paths   []string
	mu      sync.RWMutex
}

func NewTrash() *Trash {
	return &Trash{}
}

func (t *Trash) MoveToTrash(path string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.failure != nil {
		return t.failure
	}
	t.paths = append(t.paths, path)
	return nil
}

func (t *Trash) Fail(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failure = errors.New(message)
}

func (t *Trash) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.paths)
}

func (t *Trash) Item(index int) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if index < 1 || index > len(t.paths) {
		return "", false
	}
	return t.paths[index-1], true
}

func (t *Trash) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.paths = nil
	t.failure = nil
}
