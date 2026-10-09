package usecase

import "context"

type WatchWorkspace interface {
	Watch(ctx context.Context, localId, root string) error
	Unwatch(ctx context.Context) error
}
