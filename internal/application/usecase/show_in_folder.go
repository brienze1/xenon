package usecase

import "context"

type ShowInFolder interface {
	ShowInFolder(ctx context.Context, path string) error
}
