package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type clipboardWebservice struct {
	clipboard adapter.Clipboard
}

func NewClipboardWebservice(clipboard adapter.Clipboard) adapter.ClipboardWebservice {
	return &clipboardWebservice{clipboard: clipboard}
}

func (c *clipboardWebservice) Copy(ctx context.Context, text string) error {
	logger.Debug(ctx, "Started")

	if err := c.clipboard.SetText(text); err != nil {
		return errs.InternalServerError(err, "VLT-31500", "the value could not be copied to the clipboard")
	}

	logger.Debug(ctx, "Finished")
	return nil
}
