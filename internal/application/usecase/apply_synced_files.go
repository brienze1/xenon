package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ApplySyncedFiles interface {
	ApplySyncedFiles(ctx context.Context, root string, writes map[string][]byte, removals []string) ([]entity.RequestChange, error)
}
