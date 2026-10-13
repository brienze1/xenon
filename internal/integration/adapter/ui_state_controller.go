package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type UIStateController interface {
	Get(ctx context.Context, request dto.GetUIState) (*dto.UIState, error)
	Save(ctx context.Context, request dto.UIState) error
}
