package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageWindowStateService interface {
	Restore(ctx context.Context, screens []entity.Screen) (*entity.WindowState, error)
	Save(ctx context.Context, state entity.WindowState) error
}
