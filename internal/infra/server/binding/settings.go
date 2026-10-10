package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SettingsService struct {
	settings adapter.SettingsController
}

func NewSettingsService(settings adapter.SettingsController) *SettingsService {
	return &SettingsService{settings: settings}
}

func (s *SettingsService) Get(ctx context.Context) (*dto.AppSettings, error) {
	return s.settings.Get(ctx)
}

func (s *SettingsService) Save(ctx context.Context, request dto.AppSettings) (*dto.AppSettings, error) {
	return s.settings.Save(ctx, request)
}
