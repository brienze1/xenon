package adapter

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	watchDebounce     = 150 * time.Millisecond
	temporarySuffix   = ".xenon-tmp"
	gitDirectory      = ".git"
	watchStartFailure = "WKS-27500"
)

type fsFileWatcher struct {
	watcher  *fsnotify.Watcher
	pending  map[string]*time.Timer
	onChange func(path string, removed bool)
	root     string
	mu       sync.Mutex
}

func NewFileWatcher() adapter.FileWatcher {
	return &fsFileWatcher{pending: map[string]*time.Timer{}}
}

func (f *fsFileWatcher) Watch(_ context.Context, root string, onChange func(path string, removed bool)) error {
	if err := f.Stop(); err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errs.InternalServerError(err, watchStartFailure, "the workspace folder cannot be watched")
	}
	f.mu.Lock()
	f.watcher, f.root, f.onChange = watcher, root, onChange
	f.mu.Unlock()
	if err := f.addTree(watcher, root); err != nil {
		_ = f.Stop()
		return errs.InternalServerError(err, watchStartFailure, "the workspace folder cannot be watched")
	}
	go f.listen(watcher)
	return nil
}

func (f *fsFileWatcher) Stop() error {
	f.mu.Lock()
	watcher := f.watcher
	f.watcher, f.root, f.onChange = nil, "", nil
	for path, timer := range f.pending {
		timer.Stop()
		delete(f.pending, path)
	}
	f.mu.Unlock()
	if watcher == nil {
		return nil
	}
	return watcher.Close()
}

func (f *fsFileWatcher) listen(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, open := <-watcher.Events:
			if !open {
				return
			}
			f.handle(watcher, event)
		case err, open := <-watcher.Errors:
			if !open {
				return
			}
			f.overflow(watcher, err)
		}
	}
}

func (f *fsFileWatcher) handle(watcher *fsnotify.Watcher, event fsnotify.Event) {
	if event.Op == fsnotify.Chmod || strings.HasSuffix(event.Name, temporarySuffix) {
		return
	}
	if event.Has(fsnotify.Create) && isWatchableDirectory(event.Name) {
		_ = f.addTree(watcher, event.Name)
	}
	f.schedule(watcher, event.Name)
}

func (f *fsFileWatcher) overflow(watcher *fsnotify.Watcher, err error) {
	if !errors.Is(err, fsnotify.ErrEventOverflow) {
		return
	}
	f.mu.Lock()
	root := f.root
	f.mu.Unlock()
	f.schedule(watcher, root)
}

func (f *fsFileWatcher) schedule(watcher *fsnotify.Watcher, path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.watcher != watcher {
		return
	}
	if timer, found := f.pending[path]; found {
		timer.Reset(watchDebounce)
		return
	}
	f.pending[path] = time.AfterFunc(watchDebounce, func() { f.fire(watcher, path) })
}

func (f *fsFileWatcher) fire(watcher *fsnotify.Watcher, path string) {
	f.mu.Lock()
	delete(f.pending, path)
	onChange := f.onChange
	current := f.watcher == watcher
	f.mu.Unlock()
	if !current || onChange == nil {
		return
	}
	_, err := os.Lstat(path)
	onChange(path, errors.Is(err, fs.ErrNotExist))
}

func (f *fsFileWatcher) addTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if entry.Name() == gitDirectory {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

func isWatchableDirectory(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && info.Name() != gitDirectory
}
