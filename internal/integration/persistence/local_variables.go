package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const localVariablesFile = "variables.json"

type localVariablesPersistence struct {
	workspaceLocks sync.Map
	mutex          sync.Mutex
}

func NewLocalVariablesPersistence() adapter.LocalVariablesPersistence {
	return &localVariablesPersistence{}
}

func (l *localVariablesPersistence) Find(ctx context.Context, localId string) (*entity.LocalVariables, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	l.mutex.Lock()
	data, err := os.ReadFile(appDataPath(workspacesDataDir, localId, localVariablesFile))
	l.mutex.Unlock()
	if err != nil {
		return nil, errs.FileError(err, "VAR-03500")
	}
	var stored model.LocalVariables
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.UnprocessableEntityError("variables.json is not valid JSON", "VAR-04422")
	}
	local := stored.ToEntity()

	logger.Debug(ctx, "Finished", local)
	return &local, nil
}

func (l *localVariablesPersistence) Save(ctx context.Context, localId string, local entity.LocalVariables) error {
	logger.Debug(ctx, "Started", local, map[string]any{"localId": localId})

	if err := os.MkdirAll(appDataPath(workspacesDataDir, localId), appDataDirMode); err != nil {
		return errs.FileError(err, "VAR-05500")
	}
	data, _ := json.MarshalIndent(model.NewLocalVariables(local), "", "  ")
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if err := fileutils.WriteFileAtomic(appDataPath(workspacesDataDir, localId, localVariablesFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "VAR-06500")
	}

	logger.Debug(ctx, "Finished", local, map[string]any{"localId": localId})
	return nil
}

func (l *localVariablesPersistence) Delete(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	if err := os.Remove(appDataPath(workspacesDataDir, localId, localVariablesFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.FileError(err, "VAR-14500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (l *localVariablesPersistence) LockVariables(ctx context.Context, localId string) func() {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	lock, _ := l.workspaceLocks.LoadOrStore(localId, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return mutex.Unlock
}
