package logger

import (
	"fmt"
	"os"
	"sync"
)

type rotatingFile struct {
	file     *os.File
	path     string
	maxBytes int64
	size     int64
	backups  int
	mu       sync.Mutex
}

func newRotatingFile(path string, maxBytes int64, backups int) (*rotatingFile, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &rotatingFile{file: file, path: path, maxBytes: maxBytes, size: info.Size(), backups: backups}, nil
}

func (r *rotatingFile) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size+int64(len(data)) > r.maxBytes && r.size > 0 {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	written, err := r.file.Write(data)
	r.size += int64(written)
	return written, err
}

func (r *rotatingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}

func (r *rotatingFile) rotate() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	_ = os.Remove(backupName(r.path, r.backups))
	for index := r.backups - 1; index >= 1; index-- {
		_ = os.Rename(backupName(r.path, index), backupName(r.path, index+1))
	}
	if err := os.Rename(r.path, backupName(r.path, 1)); err != nil {
		return err
	}
	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	r.file = file
	r.size = 0
	return nil
}

func backupName(path string, index int) string {
	return fmt.Sprintf("%s.%d", path, index)
}
