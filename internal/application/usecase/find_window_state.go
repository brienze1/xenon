package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindWindowState interface {
	FindWindow(ctx context.Context) (*entity.WindowState, error)
}
