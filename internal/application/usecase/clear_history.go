package usecase

import "context"

type ClearHistory interface {
	Clear(ctx context.Context, localId string) error
}
