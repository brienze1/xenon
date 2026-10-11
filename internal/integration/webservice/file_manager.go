package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type fileManagerWebservice struct {
	fileManager adapter.FileManager
}

func NewFileManagerWebservice(fileManager adapter.FileManager) adapter.FileManagerWebservice {
	return &fileManagerWebservice{fileManager: fileManager}
}

func (f *fileManagerWebservice) ShowInFolder(ctx context.Context, path string) error {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	if err := f.fileManager.ShowInFolder(path); err != nil {
		return errs.InternalServerError(err, "EXP-09500", "the file manager could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": path})
	return nil
}
