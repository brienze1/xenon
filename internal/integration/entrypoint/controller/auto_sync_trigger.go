package controller

import (
	"context"

	"github.com/brienze1/xenon/pkg/logger"
)

func triggerAutoSync(ctx context.Context, trigger func(ctx context.Context, localId string) error, localId string) {
	if err := trigger(ctx, localId); err != nil {
		logger.Warn(ctx, err, "Automatic sync could not be scheduled", map[string]any{"localId": localId})
	}
}
