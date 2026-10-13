package adapter

import "context"

type WorkspaceWatcherConsumerWebservice interface {
	Consume(ctx context.Context, consumer func(localId, path string))
	Dispatch(ctx context.Context, localId, path string)
}
