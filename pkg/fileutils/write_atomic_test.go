package fileutils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestWriteFileAtomicCreatesNewFileWithMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request.yaml")
	if err := WriteFileAtomic(path, []byte("name: a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertContent(t, path, "name: a\n")
	if runtime.GOOS != "windows" {
		assertMode(t, path, 0o600)
	}
	assertNoTemp(t, path)
}

func TestWriteFileAtomicReplacesContentAndKeepsPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request.yaml")
	if err := os.WriteFile(path, []byte("old content that is longer\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertContent(t, path, "new\n")
	if runtime.GOOS != "windows" {
		assertMode(t, path, 0o640)
	}
	assertNoTemp(t, path)
}

func TestWriteFileAtomicFailsWhenDirectoryIsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "request.yaml")
	if err := WriteFileAtomic(path, []byte("x"), 0o600); err == nil {
		t.Fatal("expected an error")
	}
	assertNoTemp(t, path)
}

func TestWriteFileAtomicCleansUpWhenRenameFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request.yaml")
	failure := errors.New("rename failed")
	stubRename(t, func(string, string) error { return failure })
	if err := WriteFileAtomic(path, []byte("x"), 0o600); !errors.Is(err, failure) {
		t.Fatalf("err = %v, want %v", err, failure)
	}
	assertNoTemp(t, path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("target must not exist after a failed rename")
	}
}

func TestRenameWithRetryRetriesUntilSuccess(t *testing.T) {
	calls := 0
	stubRename(t, func(string, string) error {
		calls++
		if calls < 3 {
			return errors.New("sharing violation")
		}
		return nil
	})
	if err := renameWithRetry("a", "b", 5, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestRenameWithRetryGivesUpAfterAttempts(t *testing.T) {
	calls := 0
	stubRename(t, func(string, string) error {
		calls++
		return errors.New("sharing violation")
	})
	if err := renameWithRetry("a", "b", 5, time.Millisecond); err == nil {
		t.Fatal("expected an error")
	}
	if calls != 5 {
		t.Errorf("calls = %d, want 5", calls)
	}
}

func TestRenameAttempts(t *testing.T) {
	if renameAttempts("windows") != 5 || renameAttempts("darwin") != 1 || renameAttempts("linux") != 1 {
		t.Error("windows retries 5 times, other systems once")
	}
}

func TestTempPath(t *testing.T) {
	got := TempPath(filepath.Join("dir", "request.yaml"))
	if got != filepath.Join("dir", ".request.yaml.xenon-tmp") {
		t.Errorf("TempPath() = %q", got)
	}
}

func stubRename(t *testing.T, fn func(string, string) error) {
	t.Helper()
	original := rename
	rename = fn
	t.Cleanup(func() { rename = original })
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Errorf("content = %q, want %q", content, want)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != want {
		t.Errorf("mode = %o, want %o", info.Mode().Perm(), want)
	}
}

func assertNoTemp(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(TempPath(path)); !os.IsNotExist(err) {
		t.Error("temp file was left behind")
	}
}

func TestWriteStreamAtomicLeavesNothingWhenTheWriterFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.zip")
	failure := errors.New("cancelled")
	err := WriteStreamAtomic(path, 0o644, func(writer io.Writer) error {
		if _, err := writer.Write([]byte("partial")); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("the target exists: %v", statErr)
	}
	assertNoTemp(t, path)
}

func TestWriteStreamAtomicWritesTheStream(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.json")
	if err := WriteStreamAtomic(path, 0o600, func(writer io.Writer) error {
		_, err := io.WriteString(writer, "streamed")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	assertContent(t, path, "streamed")
	assertNoTemp(t, path)
}
