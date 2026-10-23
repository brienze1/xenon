package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type dialogWebservice struct {
	dialog adapter.Dialog
}

func NewDialogWebservice(dialog adapter.Dialog) adapter.DialogWebservice {
	return &dialogWebservice{dialog: dialog}
}

func (d *dialogWebservice) PickFolder(ctx context.Context, title string) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"title": title})

	path, err := d.dialog.PickFolder(title)
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-00500", "the folder picker could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": title, "path": path})
	return path, nil
}

func (d *dialogWebservice) PickFile(ctx context.Context, title string, filters []entity.FileFilter) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"title": title, "filters": filters})

	path, err := d.dialog.PickFile(title, filters)
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-01500", "the file picker could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": title, "filters": filters, "path": path})
	return path, nil
}

func (d *dialogWebservice) PickSources(ctx context.Context, title string) ([]string, error) {
	logger.Debug(ctx, "Started", map[string]any{"title": title})

	paths, err := d.dialog.PickSources(title)
	if err != nil {
		return nil, errs.InternalServerError(err, "DLG-03500", "the import picker could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": title, "paths": paths})
	return paths, nil
}

func (d *dialogWebservice) PickSaveFile(ctx context.Context, title, suggestedName string) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"title": title, "suggestedName": suggestedName})

	path, err := d.dialog.PickSaveFile(title, suggestedName)
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-02500", "the save dialog could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": title, "suggestedName": suggestedName, "path": path})
	return path, nil
}
