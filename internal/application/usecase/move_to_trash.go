package usecase

import "context"

type MoveToTrash interface {
	MoveToTrash(ctx context.Context, path string) error
}
