package mock

import (
	"errors"
	"sync"
)

type FileManager struct {
	failure error
	shown   []string
	mu      sync.RWMutex
}

func NewFileManager() *FileManager {
	return &FileManager{}
}

func (f *FileManager) ShowInFolder(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failure != nil {
		return f.failure
	}
	f.shown = append(f.shown, path)
	return nil
}

func (f *FileManager) Fail(message string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failure = errors.New(message)
}

func (f *FileManager) Shown() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]string{}, f.shown...)
}

func (f *FileManager) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failure = nil
	f.shown = nil
}
