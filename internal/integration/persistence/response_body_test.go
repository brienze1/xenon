package persistence

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const deadProcessId = 2147483000

func textResponse(size int) entity.Response {
	return entity.NewResponse(200, "OK", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "text/plain")}, bytes.Repeat([]byte("x"), size), int64(size), entity.Timings{})
}

func expectCode(t *testing.T, err error, code string) {
	t.Helper()
	custom, isCustom := err.(errs.BaseError)
	if !isCustom || custom.Code() != code {
		t.Fatalf("error = %v, want code %s", err, code)
	}
}

func readBody(t *testing.T, store *responseBodyPersistence, tabId string) []byte {
	t.Helper()
	body, err := store.OpenBody(context.Background(), tabId)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = body.Close() }()
	content, _ := io.ReadAll(body)
	return content
}

func useTempRoot(t *testing.T) string {
	root := t.TempDir()
	t.Setenv("TMPDIR", root)
	return root
}

func newStore(t *testing.T) (*responseBodyPersistence, string) {
	root := useTempRoot(t)
	return NewResponseBodyPersistence().(*responseBodyPersistence), filepath.Join(root, "xenon-"+strconv.Itoa(os.Getpid()))
}

func TestStoreKeepsSmallBodiesInMemory(t *testing.T) {
	store, dir := newStore(t)
	ctx := context.Background()
	if err := store.Save(ctx, "tab", textResponse(entity.MemoryBodyLimit)); err != nil {
		t.Fatal(err)
	}
	found, err := store.Find(ctx, "tab")
	if err != nil || len(found.Body) != entity.MemoryBodyLimit {
		t.Fatalf("find = %v %v", found, err)
	}
	if len(readBody(t, store, "tab")) != entity.MemoryBodyLimit {
		t.Fatal("open body must return the bytes")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("no temp folder for bodies kept in memory")
	}
}

func TestStoreWritesLargeBodiesToOwnerOnlyTempFiles(t *testing.T) {
	store, dir := newStore(t)
	ctx := context.Background()
	if err := store.Save(ctx, "tab", textResponse(entity.MemoryBodyLimit+1)); err != nil {
		t.Fatal(err)
	}
	found, _ := store.Find(ctx, "tab")
	if found.Body != nil || found.BodyLength != entity.MemoryBodyLimit+1 {
		t.Fatalf("the stored response keeps no bytes in memory: %d", len(found.Body))
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatalf("temp files = %v", files)
	}
	if runtime.GOOS != "windows" {
		dirInfo, _ := os.Stat(dir)
		fileInfo, _ := os.Stat(filepath.Join(dir, files[0].Name()))
		if dirInfo.Mode().Perm() != 0o700 || fileInfo.Mode().Perm() != 0o600 {
			t.Fatalf("permissions dir %o file %o", dirInfo.Mode().Perm(), fileInfo.Mode().Perm())
		}
	}
	if len(readBody(t, store, "tab")) != entity.MemoryBodyLimit+1 {
		t.Fatal("open body must stream the temp file")
	}

	if err := store.Save(ctx, "tab", textResponse(entity.MemoryBodyLimit+2)); err != nil {
		t.Fatal(err)
	}
	if files, _ := os.ReadDir(dir); len(files) != 1 {
		t.Fatalf("a new response replaces the old temp file, got %d files", len(files))
	}
	if err := store.Release(ctx, "tab"); err != nil {
		t.Fatal(err)
	}
	if files, _ := os.ReadDir(dir); len(files) != 0 {
		t.Fatal("release removes the temp file")
	}
	_, err := store.Find(ctx, "tab")
	expectCode(t, err, "SND-03404")
}

func TestStoreExport(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	target := filepath.Join(t.TempDir(), "out", "body.bin")
	expectCode(t, store.Export(ctx, "tab", target), "SND-03404")
	for _, size := range []int{10, entity.MemoryBodyLimit + 1} {
		if err := store.Save(ctx, "tab", textResponse(size)); err != nil {
			t.Fatal(err)
		}
		if err := store.Export(ctx, "tab", target); err != nil {
			t.Fatal(err)
		}
		content, _ := os.ReadFile(target)
		if len(content) != size {
			t.Fatalf("exported %d bytes, want %d", len(content), size)
		}
	}
	blocker := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(blocker, []byte("x"), 0o600)
	expectCode(t, store.Export(ctx, "tab", filepath.Join(blocker, "body.bin")), "SND-06500")
	expectCode(t, store.Export(ctx, "tab", t.TempDir()), "SND-07500")
}

func TestStoreFailures(t *testing.T) {
	store, blocker := newStore(t)
	_ = os.WriteFile(blocker, []byte("x"), 0o600)
	ctx := context.Background()
	expectCode(t, store.Save(ctx, "tab", textResponse(entity.MemoryBodyLimit+1)), "SND-09500")
	_, err := store.OpenBody(ctx, "none")
	expectCode(t, err, "SND-03404")

	fileStore, dir := newStore(t)
	_ = fileStore.Save(ctx, "tab", textResponse(entity.MemoryBodyLimit+1))
	_ = os.RemoveAll(dir)
	_, err = fileStore.OpenBody(ctx, "tab")
	expectCode(t, err, "SND-13500")
	expectCode(t, fileStore.Export(ctx, "tab", filepath.Join(t.TempDir(), "x")), "SND-13500")
}

func TestStoreReleaseAll(t *testing.T) {
	store, dir := newStore(t)
	ctx := context.Background()
	_ = store.Save(ctx, "a", textResponse(1))
	_ = store.Save(ctx, "b", textResponse(entity.MemoryBodyLimit+1))
	if err := store.Release(ctx, "unknown"); err != nil {
		t.Fatal(err)
	}
	if err := store.ReleaseAll(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("release all removes the temp folder")
	}
	_, err := store.Find(ctx, "a")
	expectCode(t, err, "SND-03404")
}

func TestStaleTempDirsAreRemovedOnFirstSave(t *testing.T) {
	root := useTempRoot(t)
	own := os.Getpid()
	keep := []string{
		"xenon-" + strconv.Itoa(own),
		"xenon-" + strconv.Itoa(os.Getppid()),
		"xenon-bdd-123",
		strconv.Itoa(deadProcessId),
	}
	remove := []string{"xenon-" + strconv.Itoa(deadProcessId)}
	for _, name := range append(keep, remove...) {
		_ = os.MkdirAll(filepath.Join(root, name), 0o700)
	}
	deadFile := "xenon-" + strconv.Itoa(deadProcessId+1)
	_ = os.WriteFile(filepath.Join(root, deadFile), []byte("x"), 0o600)
	if err := NewResponseBodyPersistence().Save(context.Background(), "tab", textResponse(1)); err != nil {
		t.Fatal(err)
	}
	for _, name := range append(keep, deadFile) {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Errorf("%s must be kept: %v", name, err)
		}
	}
	for _, name := range remove {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("%s must be removed", name)
		}
	}
}

func TestIsProcessRunning(t *testing.T) {
	if !isProcessRunning(os.Getpid()) || !isProcessRunning(os.Getppid()) {
		t.Fatal("this process and its parent are running")
	}
	if isProcessRunning(deadProcessId) {
		t.Fatal("an unused process id is not running")
	}
}
