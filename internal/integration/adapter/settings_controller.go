package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SettingsController interface {
	Get(ctx context.Context) (*dto.AppSettings, error)
	Save(ctx context.Context, request dto.AppSettings) (*dto.AppSettings, error)
}
