package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type pickPath struct {
	picker usecase.PickPath
}

func NewPickPathService(picker usecase.PickPath) adapter.PickPathService {
	return &pickPath{picker: picker}
}

func (p *pickPath) PickFolder(ctx context.Context, title string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"title": title})

	path, err := p.picker.PickFolder(ctx, title)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"title": title, "path": path})
	return path, nil
}

func (p *pickPath) PickFile(ctx context.Context, title string, filters []entity.FileFilter) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"title": title, "filters": filters})

	path, err := p.picker.PickFile(ctx, title, filters)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"title": title, "filters": filters, "path": path})
	return path, nil
}

func (p *pickPath) PickSaveFile(ctx context.Context, title, suggestedName string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"title": title, "suggestedName": suggestedName})

	path, err := p.picker.PickSaveFile(ctx, title, suggestedName)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"title": title, "suggestedName": suggestedName, "path": path})
	return path, nil
}
