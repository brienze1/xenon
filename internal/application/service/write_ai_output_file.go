package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

type writeAIOutputFile struct {
	aiOutputFile usecase.WriteAIOutputFile
}

func NewWriteAIOutputFileService(aiOutputFile usecase.WriteAIOutputFile) adapter.WriteAIOutputFileService {
	return &writeAIOutputFile{aiOutputFile: aiOutputFile}
}

func (w *writeAIOutputFile) Write(ctx context.Context, path string, content []byte) error {
	logger.Info(ctx, "Started", map[string]any{"path": path, "bytes": len(content)})

	if err := w.aiOutputFile.Write(ctx, path, content); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"path": path, "bytes": len(content)})
	return nil
}
