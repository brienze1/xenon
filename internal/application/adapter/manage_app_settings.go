package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAppSettingsService interface {
	Get(ctx context.Context) (*entity.AppSettings, error)
	// Save returns the saved settings even when only publishing the AI access status failed.
	Save(ctx context.Context, patch entity.AppSettingsPatch) (*entity.AppSettings, error)
}
