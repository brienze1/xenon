package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type registryPersistence struct{}

func NewRegistryPersistence() adapter.RegistryPersistence {
	return &registryPersistence{}
}

func (r *registryPersistence) Find(ctx context.Context) (*entity.AppRegistry, error) {
	logger.Debug(ctx, "Started")

	stored, err := r.read()
	if err != nil {
		return nil, err
	}
	registry := stored.ToEntity()
	for index := range registry.Workspaces {
		registry.Workspaces[index].Missing = !isDirectory(registry.Workspaces[index].Path)
	}

	logger.Debug(ctx, "Finished", registry)
	return &registry, nil
}

func (r *registryPersistence) FindWorkspace(ctx context.Context, localId string) (*entity.RegisteredWorkspace, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	registry, err := r.Find(ctx)
	if err != nil {
		return nil, err
	}
	workspace, found := registry.Find(localId)
	if !found {
		return nil, errs.NotFoundError("workspace is not registered", "WKS-07404")
	}

	logger.Debug(ctx, "Finished", workspace)
	return workspace, nil
}

func (r *registryPersistence) Save(ctx context.Context, registry entity.AppRegistry) error {
	logger.Debug(ctx, "Started", registry)

	stored, err := r.read()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(appDataPath(), appDataDirMode); err != nil {
		return errs.FileError(err, "WKS-13500")
	}
	data, _ := json.MarshalIndent(model.NewRegistry(registry, *stored), "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(registryFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "WKS-14500")
	}

	logger.Debug(ctx, "Finished", registry)
	return nil
}

func (r *registryPersistence) FindWindow(ctx context.Context) (*entity.WindowState, error) {
	logger.Debug(ctx, "Started")

	stored, err := r.read()
	if err != nil {
		return nil, err
	}
	if len(stored.Window) == 0 {
		return nil, errs.NotFoundError("window state is not saved", "WIN-04404")
	}
	var window model.WindowState
	if err := json.Unmarshal(stored.Window, &window); err != nil {
		return nil, errs.UnprocessableEntityError("window state is not valid JSON", "WIN-05422")
	}
	state := window.ToEntity()

	logger.Debug(ctx, "Finished", state)
	return &state, nil
}

func (r *registryPersistence) SaveWindow(ctx context.Context, state entity.WindowState) error {
	logger.Debug(ctx, "Started", state)

	stored, err := r.read()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(appDataPath(), appDataDirMode); err != nil {
		return errs.FileError(err, "WIN-02500")
	}
	window, _ := json.Marshal(model.NewWindowState(state))
	updated := *stored
	updated.Window = window
	data, _ := json.MarshalIndent(model.NewRegistry(stored.ToEntity(), updated), "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(registryFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "WIN-03500")
	}

	logger.Debug(ctx, "Finished", state)
	return nil
}

func (r *registryPersistence) read() (*model.Registry, error) {
	var stored model.Registry
	data, err := os.ReadFile(appDataPath(registryFile))
	if errors.Is(err, fs.ErrNotExist) {
		return &stored, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "WKS-11500")
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.InternalServerError(err, "WKS-12500", "app.json is not valid JSON")
	}
	return &stored, nil
}
