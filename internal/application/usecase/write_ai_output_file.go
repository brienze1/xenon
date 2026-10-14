package usecase

import "context"

type WriteAIOutputFile interface {
	Write(ctx context.Context, path string, content []byte) error
}
