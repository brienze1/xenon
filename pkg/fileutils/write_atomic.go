package fileutils

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	tempSuffix           = ".xenon-tmp"
	windowsRenameTries   = 5
	windowsRenameBackoff = 20 * time.Millisecond
)

var rename = os.Rename

// WriteFileAtomic replaces path with data via a temp file in the same directory; an existing file keeps its permission bits.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	return WriteStreamAtomic(path, perm, func(writer io.Writer) error {
		_, err := writer.Write(data)
		return err
	})
}

// WriteStreamAtomic is WriteFileAtomic for content produced by write; when write fails nothing is replaced and the temp file is removed.
func WriteStreamAtomic(path string, perm fs.FileMode, write func(io.Writer) error) error {
	mode := perm
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tempPath := TempPath(path)
	if err := writeTemp(tempPath, mode, write); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := renameWithRetry(tempPath, path, renameAttempts(runtime.GOOS), windowsRenameBackoff); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

// TempPath returns the temp file name WriteFileAtomic uses for path (.<name>.xenon-tmp).
func TempPath(path string) string {
	return filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+tempSuffix)
}

func writeTemp(tempPath string, mode fs.FileMode, write func(io.Writer) error) error {
	file, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if err := write(file); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func renameAttempts(goos string) int {
	if goos == "windows" {
		return windowsRenameTries
	}
	return 1
}

func renameWithRetry(from, to string, attempts int, backoff time.Duration) error {
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = rename(from, to); err == nil {
			return nil
		}
		if attempt < attempts {
			time.Sleep(backoff)
		}
	}
	return err
}
