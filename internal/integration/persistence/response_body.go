package persistence

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	tempDirPrefix     = "xenon-"
	tempDirMode       = 0o700
	tempFileMode      = 0o600
	exportDirMode     = 0o755
	exportFileMode    = 0o644
	tempBodyPattern   = "body-*"
	noResponseMessage = "the tab has no response"
)

type storedResponse struct {
	response entity.Response
	bodyPath string
}

func (s storedResponse) hasBodyFile() bool {
	return s.bodyPath != ""
}

type responseBodyPersistence struct {
	entries     map[string]storedResponse
	mu          sync.RWMutex
	cleanupOnce sync.Once
}

func NewResponseBodyPersistence() adapter.ResponseBodyPersistence {
	return &responseBodyPersistence{entries: make(map[string]storedResponse)}
}

func (r *responseBodyPersistence) Save(ctx context.Context, tabId string, response entity.Response) error {
	logger.Debug(ctx, "Started", response, map[string]any{"tabId": tabId})

	r.cleanupOnce.Do(removeStaleDirs)
	stored := storedResponse{response: response}
	if !response.KeepsBodyInMemory() {
		path, err := r.writeTempBody(response.Body)
		if err != nil {
			return err
		}
		stored = storedResponse{response: response.WithoutBody(), bodyPath: path}
	}
	r.mu.Lock()
	previous := r.entries[tabId]
	r.entries[tabId] = stored
	r.mu.Unlock()
	removeBodyFile(previous)

	logger.Debug(ctx, "Finished", response, map[string]any{"tabId": tabId})
	return nil
}

func (r *responseBodyPersistence) Find(ctx context.Context, tabId string) (*entity.Response, error) {
	logger.Debug(ctx, "Started", map[string]any{"tabId": tabId})

	stored, err := r.entry(tabId)
	if err != nil {
		return nil, err
	}

	response := stored.response

	logger.Debug(ctx, "Finished", response, map[string]any{"tabId": tabId})
	return &response, nil
}

func (r *responseBodyPersistence) OpenBody(ctx context.Context, tabId string) (io.ReadCloser, error) {
	logger.Debug(ctx, "Started", map[string]any{"tabId": tabId})

	stored, err := r.entry(tabId)
	if err != nil {
		return nil, err
	}
	body, err := openBody(stored)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"tabId": tabId})
	return body, nil
}

func (r *responseBodyPersistence) Export(ctx context.Context, tabId, path string) error {
	logger.Debug(ctx, "Started", map[string]any{"tabId": tabId, "path": path})

	stored, err := r.entry(tabId)
	if err != nil {
		return err
	}
	body, err := openBody(stored)
	if err != nil {
		return err
	}
	defer func() { _ = body.Close() }()
	content, err := io.ReadAll(body)
	if err != nil {
		return errs.FileError(err, "SND-05500")
	}
	if err := os.MkdirAll(filepath.Dir(path), exportDirMode); err != nil {
		return errs.FileError(err, "SND-06500")
	}
	if err := fileutils.WriteFileAtomic(path, content, exportFileMode); err != nil {
		return errs.FileError(err, "SND-07500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"tabId": tabId, "path": path})
	return nil
}

func (r *responseBodyPersistence) Release(ctx context.Context, tabId string) error {
	logger.Debug(ctx, "Started", map[string]any{"tabId": tabId})

	r.mu.Lock()
	stored := r.entries[tabId]
	delete(r.entries, tabId)
	r.mu.Unlock()
	removeBodyFile(stored)

	logger.Debug(ctx, "Finished", map[string]any{"tabId": tabId})
	return nil
}

func (r *responseBodyPersistence) ReleaseAll(ctx context.Context) error {
	logger.Debug(ctx, "Started")

	r.mu.Lock()
	r.entries = make(map[string]storedResponse)
	r.mu.Unlock()
	if err := os.RemoveAll(ownTempDir()); err != nil {
		return errs.FileError(err, "SND-08500")
	}

	logger.Debug(ctx, "Finished")
	return nil
}

func (r *responseBodyPersistence) entry(tabId string) (storedResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stored, found := r.entries[tabId]
	if !found {
		return storedResponse{}, errs.NotFoundError(noResponseMessage, "SND-03404")
	}
	return stored, nil
}

func (r *responseBodyPersistence) writeTempBody(body []byte) (string, error) {
	dir := ownTempDir()
	if err := os.MkdirAll(dir, tempDirMode); err != nil {
		return "", errs.FileError(err, "SND-09500")
	}
	file, err := os.CreateTemp(dir, tempBodyPattern)
	if err != nil {
		return "", errs.FileError(err, "SND-10500")
	}
	defer func() { _ = file.Close() }()
	if err := file.Chmod(tempFileMode); err != nil {
		return "", errs.FileError(err, "SND-11500")
	}
	if _, err := file.Write(body); err != nil {
		return "", errs.FileError(err, "SND-12500")
	}
	return file.Name(), nil
}

func openBody(stored storedResponse) (io.ReadCloser, error) {
	if !stored.hasBodyFile() {
		return io.NopCloser(bytes.NewReader(stored.response.Body)), nil
	}
	file, err := os.Open(stored.bodyPath)
	if err != nil {
		return nil, errs.FileError(err, "SND-13500")
	}
	return file, nil
}

func removeBodyFile(stored storedResponse) {
	_ = os.Remove(stored.bodyPath)
}

func ownTempDir() string {
	return filepath.Join(properties.Properties().Application.TempDir, tempDirPrefix+strconv.Itoa(os.Getpid()))
}

func removeStaleDirs() {
	tempRoot := properties.Properties().Application.TempDir
	processId := os.Getpid()
	entries, _ := os.ReadDir(tempRoot)
	for _, entry := range entries {
		if isStaleDir(entry, processId) {
			_ = os.RemoveAll(filepath.Join(tempRoot, entry.Name()))
		}
	}
}

func isStaleDir(entry os.DirEntry, processId int) bool {
	owner, isOwned := strings.CutPrefix(entry.Name(), tempDirPrefix)
	ownerId, err := strconv.Atoi(owner)
	return entry.IsDir() && isOwned && err == nil && ownerId != processId && !isProcessRunning(ownerId)
}
