package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type FileChangeController interface {
	Handle(ctx context.Context, request dto.FileChange) error
}
