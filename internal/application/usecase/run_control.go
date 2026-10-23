package usecase

import "context"

type RunControl interface {
	Pause()
	Resume()
	Stop()
	Abandon()
	AwaitFinished(ctx context.Context) bool
}
