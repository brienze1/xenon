package usecase

import "context"

type ReleaseActiveRun interface {
	Release(ctx context.Context, localId, runId string)
}
