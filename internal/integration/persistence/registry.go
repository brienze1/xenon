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
	data, _ := json.MarshalIndent(model.NewRegistry(registry, stored.Window), "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(registryFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, "WKS-14500")
	}

	logger.Debug(ctx, "Finished", registry)
	return nil
}

func (r *registryPersistence) read() (*model.Registry, error) {
	data, err := os.ReadFile(appDataPath(registryFile))
	if errors.Is(err, fs.ErrNotExist) {
		return &model.Registry{}, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "WKS-11500")
	}
	var stored model.Registry
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.InternalServerError(err, "WKS-12500", "app.json is not valid JSON")
	}
	return &stored, nil
}
