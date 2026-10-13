package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type windowFocusWebservice struct {
	windowFocuser adapter.WindowFocuser
}

func NewWindowFocusWebservice(windowFocuser adapter.WindowFocuser) adapter.WindowFocusWebservice {
	return &windowFocusWebservice{windowFocuser: windowFocuser}
}

func (w *windowFocusWebservice) Focus(ctx context.Context) {
	logger.Debug(ctx, "Started")

	w.windowFocuser.Focus()

	logger.Debug(ctx, "Finished")
}
