package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageUIStateService interface {
	Get(ctx context.Context, localId string) (*entity.UIState, error)
	Save(ctx context.Context, localId string, state entity.UIState) error
}
