package usecase

import "context"

type FindIterationData interface {
	FindIterationData(ctx context.Context, path string) ([]byte, error)
	IterationDataExists(ctx context.Context, path string) bool
}
