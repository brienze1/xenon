package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAppSettingsService interface {
	Get(ctx context.Context) (*entity.AppSettings, error)
	Save(ctx context.Context, patch entity.AppSettingsPatch) (*entity.AppSettings, error)
}
