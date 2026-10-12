package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadApiDocument interface {
	ReadDocument(ctx context.Context, data []byte) (*entity.ApiDocument, error)
}
