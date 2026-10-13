package usecase

import "context"

type DeleteSshKey interface {
	DeleteKey(ctx context.Context, path string) error
}
