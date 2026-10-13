package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadCurlCommand interface {
	ReadCommand(ctx context.Context, text string) (*entity.ParsedCurl, error)
}
