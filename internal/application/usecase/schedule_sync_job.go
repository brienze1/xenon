package usecase

import (
	"context"
	"time"
)

type ScheduleSyncJob interface {
	// Schedule replaces the job pending under the same key; the job runs detached from ctx's cancellation.
	Schedule(ctx context.Context, key string, delay time.Duration, job func(ctx context.Context))
}
