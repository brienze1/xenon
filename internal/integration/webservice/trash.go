package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type trashWebservice struct {
	trash adapter.Trash
}

func NewTrashWebservice(trash adapter.Trash) adapter.TrashWebservice {
	return &trashWebservice{trash: trash}
}

func (t *trashWebservice) MoveToTrash(ctx context.Context, path string) error {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	if err := t.trash.MoveToTrash(path); err != nil {
		return errs.InternalServerError(err, "WKS-05500", "the folder could not be moved to the trash; remove it from Xenon and keep the files instead")
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": path})
	return nil
}
