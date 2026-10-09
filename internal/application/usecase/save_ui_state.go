package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveUIState interface {
	Save(ctx context.Context, localId string, state entity.UIState) error
}
