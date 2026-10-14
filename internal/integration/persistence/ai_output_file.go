package persistence

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	aiOutputFileMode       = 0o644
	aiOutputFileUnwritable = "Invalid arguments for export: the body could not be written to "
)

type aiOutputFilePersistence struct{}

func NewAIOutputFilePersistence() adapter.AIOutputFilePersistence {
	return &aiOutputFilePersistence{}
}

func (a *aiOutputFilePersistence) Write(ctx context.Context, path string, content []byte) error {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "bytes": len(content)})

	if err := fileutils.WriteFileAtomic(path, content, aiOutputFileMode); err != nil {
		logger.Error(ctx, err, "The AI output file could not be written", map[string]any{"path": path})
		return errs.BadRequestError(aiOutputFileUnwritable+path, entity.AIInvalidArgumentsCode)
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "bytes": len(content)})
	return nil
}
