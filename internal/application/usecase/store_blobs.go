package usecase

import "context"

type StoreBlobs interface {
	StoreWorktreeBlobs(ctx context.Context, clonePath, subfolder string, paths []string) error
	StoreBlob(ctx context.Context, clonePath string, content []byte) (string, error)
}
