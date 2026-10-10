package usecase

import "context"

type DiscardImportSource interface {
	Discard(ctx context.Context, id string) error
	DiscardFiles(ctx context.Context, dirs []string) error
}
