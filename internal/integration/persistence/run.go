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
	runsDataDir   = "runs"
	runFileSuffix = ".json"
)

type runPersistence struct {
	workspaceLocks sync.Map
}

func NewRunPersistence() adapter.RunPersistence {
	return &runPersistence{}
}

func (r *runPersistence) SaveRun(ctx context.Context, run entity.Run) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID, "status": run.Status})

	unlock := r.lock(run.Target.LocalID)
	defer unlock()
	folder := appDataPath(workspacesDataDir, run.Target.LocalID, runsDataDir, run.Target.ItemID)
	if err := os.MkdirAll(folder, appDataDirMode); err != nil {
		return errs.FileError(err, "RUN-13500")
	}
	data, err := json.MarshalIndent(model.NewRun(run), "", "  ")
	if err != nil {
		return errs.FileError(err, "RUN-14500")
	}
	if err := fileutils.WriteFileAtomic(filepath.Join(folder, run.ID+runFileSuffix), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "RUN-15500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID})
	return nil
}

func (r *runPersistence) FindRun(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	if strings.ContainsAny(runId, `/\*?[`) {
		return nil, usecase.RejectUnknownRun(runId)
	}
	unlock := r.lock(localId)
	defer unlock()
	paths, err := filepath.Glob(appDataPath(workspacesDataDir, localId, runsDataDir, "*", runId+runFileSuffix))
	if err != nil || len(paths) == 0 {
		return nil, usecase.RejectUnknownRun(runId)
	}
	run, err := readRun(paths[0], localId)
	if err != nil {
		return nil, usecase.RejectUnknownRun(runId)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (r *runPersistence) FindRuns(ctx context.Context, localId, targetId string) ([]entity.Run, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "targetId": targetId})

	unlock := r.lock(localId)
	defer unlock()
	folder := appDataPath(workspacesDataDir, localId, runsDataDir, targetId)
	entries, err := os.ReadDir(folder)
	if errors.Is(err, fs.ErrNotExist) {
		return []entity.Run{}, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "RUN-16500")
	}
	runs := make([]entity.Run, 0, len(entries))
	for _, file := range slices.Backward(entries) {
		if file.IsDir() || !strings.HasSuffix(file.Name(), runFileSuffix) {
			continue
		}
		run, err := readRun(filepath.Join(folder, file.Name()), localId)
		if err != nil {
			logger.Warn(ctx, err, "Skipping an unreadable run file", map[string]any{"localId": localId, "file": file.Name()})
			continue
		}
		runs = append(runs, *run)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "targetId": targetId, "runs": len(runs)})
	return runs, nil
}

func (r *runPersistence) lock(localId string) func() {
	lock, _ := r.workspaceLocks.LoadOrStore(localId, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
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
