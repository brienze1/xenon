package mock

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
)

type FileWatcher struct {
	onChange func(path string, removed bool)
	root     string
	mu       sync.RWMutex
}

func NewFileWatcher() *FileWatcher {
	return &FileWatcher{}
}

func (f *FileWatcher) Watch(_ context.Context, root string, onChange func(path string, removed bool)) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.root = root
	f.onChange = onChange
	return nil
}

func (f *FileWatcher) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.root = ""
	f.onChange = nil
	return nil
}

func (f *FileWatcher) Root() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.root
}

func (f *FileWatcher) Report(path string, removed bool) {
	f.mu.RLock()
	root, onChange := f.root, f.onChange
	f.mu.RUnlock()
	if onChange == nil || !isWithin(path, root) {
		return
	}
	onChange(path, removed)
}

func (f *FileWatcher) Reset() {
	_ = f.Stop()
}

func isWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
