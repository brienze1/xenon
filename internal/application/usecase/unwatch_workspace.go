package usecase

import "context"

type UnwatchWorkspace interface {
	Unwatch(ctx context.Context) error
}
