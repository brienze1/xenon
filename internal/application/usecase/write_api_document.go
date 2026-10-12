package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type WriteApiDocument interface {
	WriteDocument(ctx context.Context, document entity.ApiDocument) ([]byte, error)
}
