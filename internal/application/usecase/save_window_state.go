package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveWindowState interface {
	SaveWindow(ctx context.Context, state entity.WindowState) error
}
