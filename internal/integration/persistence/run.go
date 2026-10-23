package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	runsDataDir     = "runs"
	runFileSuffix   = ".json"
	storedRunsLimit = 25
)

type runPersistence struct {
	workspaceLocks sync.Map
}

func NewRunPersistence() adapter.RunPersistence {
	return &runPersistence{}
}

func (r *runPersistence) SaveRun(ctx context.Context, run entity.Run) ([]entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID, "status": run.Status})

	unlock := r.lock(run.Target.LocalID)
	defer unlock()
	folder := runsFolder(run.Target.LocalID, run.Target.ItemID)
	if err := os.MkdirAll(folder, appDataDirMode); err != nil {
		return nil, errs.FileError(err, "RUN-13500")
	}
	data, err := json.MarshalIndent(model.NewRun(run), "", "  ")
	if err != nil {
		return nil, errs.FileError(err, "RUN-14500")
	}
	if err := fileutils.WriteFileAtomic(filepath.Join(folder, run.ID+runFileSuffix), append(data, '\n'), appDataFileMode); err != nil {
		return nil, errs.FileError(err, "RUN-15500")
	}
	pruned, err := pruneRuns(ctx, folder, run.Target.LocalID)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID, "pruned": len(pruned)})
	return pruned, nil
}

func (r *runPersistence) FindRun(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	unlock := r.lock(localId)
	defer unlock()
	path, err := runFile(localId, runId)
	if err != nil {
		return nil, err
	}
	run, err := readRun(path, localId)
	if err != nil {
		return nil, usecase.RejectUnknownRun(runId)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (r *runPersistence) FindRunFile(ctx context.Context, localId, runId string) ([]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	unlock := r.lock(localId)
	defer unlock()
	path, err := runFile(localId, runId)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, usecase.RejectUnknownRun(runId)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "bytes": len(data)})
	return data, nil
}

func (r *runPersistence) FindRuns(ctx context.Context, localId, targetId string) ([]entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "targetId": targetId})

	unlock := r.lock(localId)
	defer unlock()
	files, err := runFiles(runsFolder(localId, targetId))
	if err != nil {
		return nil, errs.FileError(err, "RUN-16500")
	}
	runs := make([]entity.Run, 0, len(files))
	for _, path := range slices.Backward(files) {
		run, err := readRun(path, localId)
		if err != nil {
			logger.Warn(ctx, err, "Skipping an unreadable run file", map[string]any{"localId": localId, "file": filepath.Base(path)})
			continue
		}
		runs = append(runs, *run)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "targetId": targetId, "runs": len(runs)})
	return runs, nil
}

func (r *runPersistence) DeleteRun(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	unlock := r.lock(localId)
	defer unlock()
	path, err := runFile(localId, runId)
	if err != nil {
		return nil, err
	}
	deleted := storedRun(path, localId)
	if err := os.Remove(path); err != nil {
		return nil, errs.FileError(err, "RUN-17500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
	return &deleted, nil
}

func (r *runPersistence) ClearRuns(ctx context.Context, localId, targetId string) ([]entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "targetId": targetId})

	unlock := r.lock(localId)
	defer unlock()
	deleted, err := deleteRunsFolder(runsFolder(localId, targetId), localId, "RUN-18500")
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "targetId": targetId, "runs": len(deleted)})
	return deleted, nil
}

func (r *runPersistence) DeleteOrphanRuns(ctx context.Context, localId string, targetIds []string) ([]entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "targets": len(targetIds)})

	unlock := r.lock(localId)
	defer unlock()
	folders, err := os.ReadDir(appDataPath(workspacesDataDir, localId, runsDataDir))
	if errors.Is(err, fs.ErrNotExist) {
		return []entity.Run{}, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "RUN-19500")
	}
	deleted := make([]entity.Run, 0)
	for _, folder := range folders {
		if !folder.IsDir() || slices.Contains(targetIds, folder.Name()) {
			continue
		}
		runs, err := deleteRunsFolder(runsFolder(localId, folder.Name()), localId, "RUN-19500")
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, runs...)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runs": len(deleted)})
	return deleted, nil
}

func (r *runPersistence) DeleteWorkspaceRuns(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	unlock := r.lock(localId)
	defer unlock()
	if err := os.RemoveAll(appDataPath(workspacesDataDir, localId, runsDataDir)); err != nil {
		return errs.FileError(err, "RUN-20500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (r *runPersistence) lock(localId string) func() {
	lock, _ := r.workspaceLocks.LoadOrStore(localId, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
}

func pruneRuns(ctx context.Context, folder, localId string) ([]entity.Run, error) {
	files, err := runFiles(folder)
	if err != nil {
		return nil, errs.FileError(err, "RUN-21500")
	}
	pruned := make([]entity.Run, 0)
	for len(files) > storedRunsLimit {
		pruned = append(pruned, storedRun(files[0], localId))
		if err := os.Remove(files[0]); err != nil {
			return nil, errs.FileError(err, "RUN-21500")
		}
		logger.Debug(ctx, "Pruned a stored run past the limit", map[string]any{"localId": localId, "file": filepath.Base(files[0])})
		files = files[1:]
	}
	return pruned, nil
}

func deleteRunsFolder(folder, localId, code string) ([]entity.Run, error) {
	files, err := runFiles(folder)
	if err != nil {
		return nil, errs.FileError(err, code)
	}
	deleted := make([]entity.Run, 0, len(files))
	for _, path := range files {
		deleted = append(deleted, storedRun(path, localId))
	}
	if err := os.RemoveAll(folder); err != nil {
		return nil, errs.FileError(err, code)
	}
	return deleted, nil
}

func runFiles(folder string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, file := range entries {
		if !file.IsDir() && strings.HasSuffix(file.Name(), runFileSuffix) {
			files = append(files, filepath.Join(folder, file.Name()))
		}
	}
	return files, nil
}

func runFile(localId, runId string) (string, error) {
	if strings.ContainsAny(runId, `/\*?[`) {
		return "", usecase.RejectUnknownRun(runId)
	}
	paths, err := filepath.Glob(appDataPath(workspacesDataDir, localId, runsDataDir, "*", runId+runFileSuffix))
	if err != nil || len(paths) == 0 {
		return "", usecase.RejectUnknownRun(runId)
	}
	return paths[0], nil
}

func runsFolder(localId, targetId string) string {
	return appDataPath(workspacesDataDir, localId, runsDataDir, targetId)
}

func storedRun(path, localId string) entity.Run {
	run, err := readRun(path, localId)
	if err != nil {
		return entity.Run{ID: strings.TrimSuffix(filepath.Base(path), runFileSuffix)}
	}
	return *run
}

func readRun(path, localId string) (*entity.Run, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stored model.Run
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	run := stored.ToEntity(localId)
	return &run, nil
}
