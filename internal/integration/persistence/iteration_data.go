package persistence

import (
	"context"
	"os"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type iterationDataPersistence struct{}

func NewIterationDataPersistence() adapter.IterationDataPersistence {
	return &iterationDataPersistence{}
}

func (i *iterationDataPersistence) FindIterationData(ctx context.Context, path string) ([]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Warn(ctx, err, "The iteration data file cannot be read", map[string]any{"path": path})
		return nil, usecase.RejectInvalidIterationData("the file cannot be read")
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "bytes": len(content)})
	return content, nil
}

func (i *iterationDataPersistence) IterationDataExists(ctx context.Context, path string) bool {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	info, err := os.Stat(path)
	exists := err == nil && !info.IsDir()

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "exists": exists})
	return exists
}
