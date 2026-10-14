package adapter

import "context"

type WriteAIOutputFileService interface {
	Write(ctx context.Context, path string, content []byte) error
}
