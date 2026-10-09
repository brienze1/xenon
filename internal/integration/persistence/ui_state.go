package persistence

import (
	"context"
	"encoding/json"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	workspacesDataDir = "workspaces"
	uiStateFile       = "state.json"
)

type uiStatePersistence struct{}

func NewUIStatePersistence() adapter.UIStatePersistence {
	return &uiStatePersistence{}
}

func (u *uiStatePersistence) Find(ctx context.Context, localId string) (*entity.UIState, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	data, err := os.ReadFile(appDataPath(workspacesDataDir, localId, uiStateFile))
	if err != nil {
		return nil, errs.FileError(err, "UIS-02500")
	}
	var stored model.UIState
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.UnprocessableEntityError("state.json is not valid JSON", "UIS-06422")
	}
	state := stored.ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return &state, nil
}

func (u *uiStatePersistence) Save(ctx context.Context, localId string, state entity.UIState) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "tabs": len(state.Tabs)})

	if err := os.MkdirAll(appDataPath(workspacesDataDir, localId), appDataDirMode); err != nil {
		return errs.FileError(err, "UIS-03500")
	}
	data, _ := json.MarshalIndent(model.NewUIState(state), "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(workspacesDataDir, localId, uiStateFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "UIS-04500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return nil
}

func (u *uiStatePersistence) Delete(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	if err := os.RemoveAll(appDataPath(workspacesDataDir, localId)); err != nil {
		return errs.FileError(err, "UIS-05500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}
