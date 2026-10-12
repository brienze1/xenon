package adapter

import "context"

type AutoSyncService interface {
	Opened(ctx context.Context, localId string) error
	Saved(ctx context.Context, localId string) error
}
