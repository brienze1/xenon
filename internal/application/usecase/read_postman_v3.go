package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadPostmanV3 interface {
	ReadPostmanV3(ctx context.Context, input entity.ImportInput) (*entity.PostmanV3Source, error)
}
