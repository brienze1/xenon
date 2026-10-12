package usecase

import "context"

type LocateClone interface {
	DefaultClonePath(ctx context.Context, localId string) string
}
