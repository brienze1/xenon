package persistence

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const environmentExtension = ".yaml"

type environmentPersistence struct {
	index adapter.WorkspaceIndexCache
}

func NewEnvironmentPersistence(index adapter.WorkspaceIndexCache) adapter.EnvironmentPersistence {
	return &environmentPersistence{index: index}
}

func (e *environmentPersistence) List(ctx context.Context, root string) ([]entity.Environment, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	environments, err := readEnvironments(root)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "environments": len(environments)})
	return environments, nil
}

func (e *environmentPersistence) Find(ctx context.Context, root, id string) (*entity.Environment, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})

	environment, err := findEnvironment(root, id)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", *environment)
	return environment, nil
}

func (e *environmentPersistence) Save(ctx context.Context, root string, environment entity.Environment) (*entity.Environment, error) {
	logger.Debug(ctx, "Started", environment, map[string]any{"root": root})

	dir := filepath.Join(root, environmentsDir)
	if err := os.MkdirAll(dir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "ENV-06500")
	}
	environments, err := readEnvironments(root)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, environment.FileName)
	document := yamlnode.NewDocument()
	if data, err := os.ReadFile(path); err == nil {
		if parsed, err := yamlnode.Parse(data); err == nil {
			document = parsed
		}
	}
	stored := model.NewEnvironment(environment)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	data, err := yamlnode.Encode(document)
	if err != nil {
		return nil, errs.InternalServerError(err, "ENV-07500", "the environment file could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return nil, errs.FileError(err, "ENV-08500")
	}
	e.index.RecordWrite(ctx, path, data)
	for _, previous := range environments {
		if previous.HasID(environment.ID) && previous.FileName != environment.FileName {
			if err := os.Remove(filepath.Join(dir, previous.FileName)); err != nil {
				return nil, errs.FileError(err, "ENV-09500")
			}
		}
	}
	saved := readEnvironment(dir, environment.FileName)

	logger.Debug(ctx, "Finished", saved, map[string]any{"root": root})
	return &saved, nil
}

func (e *environmentPersistence) Delete(ctx context.Context, root, id string) error {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})

	environment, err := findEnvironment(root, id)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(root, environmentsDir, environment.FileName)); err != nil {
		return errs.FileError(err, "ENV-10500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return nil
}

func findEnvironment(root, id string) (*entity.Environment, error) {
	environments, err := readEnvironments(root)
	if err != nil {
		return nil, err
	}
	for index := range environments {
		if environments[index].HasID(id) {
			return &environments[index], nil
		}
	}
	return nil, errs.NotFoundError("the environment does not exist", "ENV-01404")
}

func readEnvironments(root string) ([]entity.Environment, error) {
	dir := filepath.Join(root, environmentsDir)
	listing, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []entity.Environment{}, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "ENV-05500")
	}
	environments := make([]entity.Environment, 0, len(listing))
	for _, entry := range listing {
		if isEnvironmentFile(entry) {
			environments = append(environments, readEnvironment(dir, entry.Name()))
		}
	}
	return environments, nil
}

func readEnvironment(dir, fileName string) entity.Environment {
	var stored model.Environment
	if reason := readDocument(filepath.Join(dir, fileName), itemkind.Environment, &stored.Xenon, &stored.Kind, &stored); reason != "" {
		return entity.NewMalformedEnvironment(fileName, reason)
	}
	if strings.TrimSpace(stored.Id) == "" {
		return entity.NewMalformedEnvironment(fileName, fileName+" has no id")
	}
	return stored.ToEntity(fileName)
}

func isEnvironmentFile(entry fs.DirEntry) bool {
	return entry.Type().IsRegular() && !isHidden(entry) && strings.EqualFold(filepath.Ext(entry.Name()), environmentExtension)
}
