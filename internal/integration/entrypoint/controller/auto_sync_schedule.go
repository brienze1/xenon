package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

func scheduleAutoSync(ctx context.Context, errorHandler adapter.ErrorHandler, trigger func(ctx context.Context, localId string) error, localId string) {
	if err := trigger(ctx, localId); err != nil {
		errorHandler.Handle(ctx, err, map[string]any{"localId": localId})
	}
}
