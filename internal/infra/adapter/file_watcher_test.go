package adapter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/brienze1/xenon/pkg/errs"
)

const (
	eventWait = 2 * time.Second
	quietWait = 600 * time.Millisecond
)

type recordedChange struct {
	path    string
	removed bool
}

type changeRecorder struct {
	changes chan recordedChange
	mu      sync.Mutex
	count   int
}

func newChangeRecorder() *changeRecorder {
	return &changeRecorder{changes: make(chan recordedChange, 64)}
}

func (c *changeRecorder) record(path string, removed bool) {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
	c.changes <- recordedChange{path: path, removed: removed}
}

func (c *changeRecorder) total() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func (c *changeRecorder) next(t *testing.T) recordedChange {
	t.Helper()
	select {
	case change := <-c.changes:
		return change
	case <-time.After(eventWait):
		t.Fatal("expected a change, got none")
		return recordedChange{}
	}
}

func (c *changeRecorder) expectNone(t *testing.T) {
	t.Helper()
	select {
	case change := <-c.changes:
		t.Fatalf("expected no change, got %+v", change)
	case <-time.After(quietWait):
	}
}

func watchedRoot(t *testing.T) (string, *fsFileWatcher, *changeRecorder) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "collections", "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	watcher := NewFileWatcher().(*fsFileWatcher)
	recorder := newChangeRecorder()
	if err := watcher.Watch(context.Background(), root, recorder.record); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = watcher.Stop() })
	return root, watcher, recorder
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWatcherReportsAWrittenFileOnceAfterABurst(t *testing.T) {
	root, _, recorder := watchedRoot(t)
	path := filepath.Join(root, "collections", "billing", "ping.yaml")
	for index := 0; index < 5; index++ {
		write(t, path, "name: ping")
		time.Sleep(20 * time.Millisecond)
	}
	change := recorder.next(t)
	if change.path != path || change.removed {
		t.Fatalf("change = %+v, want %s not removed", change, path)
	}
	recorder.expectNone(t)
}

func TestWatcherDebouncesForAtLeastTheDebounceWindow(t *testing.T) {
	root, _, recorder := watchedRoot(t)
	write(t, filepath.Join(root, "collections", "billing", "ping.yaml"), "name: ping")
	time.Sleep(watchDebounce / 2)
	if recorder.total() != 0 {
		t.Fatal("a change was reported before the debounce window ended")
	}
	recorder.next(t)
}

func TestWatcherReportsRemovals(t *testing.T) {
	root, _, recorder := watchedRoot(t)
	path := filepath.Join(root, "collections", "billing", "ping.yaml")
	write(t, path, "name: ping")
	recorder.next(t)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	change := recorder.next(t)
	if change.path != path || !change.removed {
		t.Fatalf("change = %+v, want %s removed", change, path)
	}
}

func TestWatcherFollowsNewFolders(t *testing.T) {
	root, _, recorder := watchedRoot(t)
	folder := filepath.Join(root, "collections", "billing", "invoices")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if change := recorder.next(t); change.path != folder {
		t.Fatalf("change = %+v, want the new folder", change)
	}
	path := filepath.Join(folder, "list.yaml")
	write(t, path, "name: list")
	if change := recorder.next(t); change.path != path {
		t.Fatalf("change = %+v, want %s", change, path)
	}
}

func TestWatcherIgnoresTemporaryFilesChmodAndGit(t *testing.T) {
	root, _, recorder := watchedRoot(t)
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	recorder.next(t)
	path := filepath.Join(root, "collections", "billing", "ping.yaml")
	write(t, path, "name: ping")
	recorder.next(t)
	write(t, filepath.Join(root, "collections", "billing", ".ping.yaml.xenon-tmp"), "partial")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	recorder.expectNone(t)
}

func TestWatcherSkipsGitFoldersPresentAtStart(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git", "objects"), 0o755); err != nil {
		t.Fatal(err)
	}
	watcher := NewFileWatcher().(*fsFileWatcher)
	recorder := newChangeRecorder()
	if err := watcher.Watch(context.Background(), root, recorder.record); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = watcher.Stop() })
	write(t, filepath.Join(root, ".git", "objects", "pack"), "data")
	recorder.expectNone(t)
}

func TestWatcherStopsReportingAfterStop(t *testing.T) {
	root, watcher, recorder := watchedRoot(t)
	path := filepath.Join(root, "collections", "billing", "ping.yaml")
	write(t, path, "name: ping")
	time.Sleep(watchDebounce / 3)
	if err := watcher.Stop(); err != nil {
		t.Fatal(err)
	}
	write(t, path, "name: pong")
	recorder.expectNone(t)
	if err := watcher.Stop(); err != nil {
		t.Fatalf("stopping twice must be harmless: %v", err)
	}
}

func TestWatchingAgainReplacesThePreviousRoot(t *testing.T) {
	first, watcher, recorder := watchedRoot(t)
	second, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := watcher.Watch(context.Background(), second, recorder.record); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(first, "collections", "billing", "old.yaml"), "name: old")
	recorder.expectNone(t)
	path := filepath.Join(second, "new.yaml")
	write(t, path, "name: new")
	if change := recorder.next(t); change.path != path {
		t.Fatalf("change = %+v, want %s", change, path)
	}
}

func TestWatchingAMissingFolderFails(t *testing.T) {
	watcher := NewFileWatcher()
	err := watcher.Watch(context.Background(), filepath.Join(t.TempDir(), "missing"), func(string, bool) {})
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.Code() != "WKS-29500" {
		t.Fatalf("Watch() = %v, want WKS-29500", err)
	}
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Stop() after a failed watch = %v", err)
	}
}

func TestOverflowReloadsTheWholeRoot(t *testing.T) {
	root, watcher, recorder := watchedRoot(t)
	watcher.mu.Lock()
	current := watcher.watcher
	watcher.mu.Unlock()
	watcher.overflow(current, errors.New("some other error"))
	recorder.expectNone(t)
	watcher.overflow(current, fsnotify.ErrEventOverflow)
	if change := recorder.next(t); change.path != root || change.removed {
		t.Fatalf("change = %+v, want a reload of %s", change, root)
	}
}

func TestEventsOfAReplacedWatcherAreDropped(t *testing.T) {
	root, watcher, recorder := watchedRoot(t)
	stale, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stale.Close() })
	watcher.schedule(stale, filepath.Join(root, "x.yaml"))
	watcher.fire(stale, filepath.Join(root, "x.yaml"))
	recorder.expectNone(t)
}

func TestMovingAMissingPathToTheTrashFails(t *testing.T) {
	if err := NewTrash().MoveToTrash(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("MoveToTrash() of a missing path must fail")
	}
}

func TestWatcherIgnoresImportStagingFolders(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".xenon-import-present", "item"), 0o755); err != nil {
		t.Fatal(err)
	}
	watcher := NewFileWatcher().(*fsFileWatcher)
	recorder := newChangeRecorder()
	if err := watcher.Watch(context.Background(), root, recorder.record); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = watcher.Stop() })
	write(t, filepath.Join(root, ".xenon-import-present", "item", "a.yaml"), "data")
	if err := os.MkdirAll(filepath.Join(root, ".xenon-import-new", "item"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, ".xenon-import-new", "b.yaml"), "data")
	recorder.expectNone(t)
	write(t, filepath.Join(root, "workspace.yaml"), "name: Acme")
	if change := recorder.next(t); change.path != filepath.Join(root, "workspace.yaml") {
		t.Fatalf("change = %+v", change)
	}
}
