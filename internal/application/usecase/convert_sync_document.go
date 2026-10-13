package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ConvertSyncDocument interface {
	DecodeDocument(ctx context.Context, content []byte) (*entity.DocumentNode, error)
	EncodeDocument(ctx context.Context, node *entity.DocumentNode) ([]byte, error)
}
