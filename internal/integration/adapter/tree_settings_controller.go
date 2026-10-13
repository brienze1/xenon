package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeSettingsController interface {
	GetSettings(ctx context.Context, request dto.TreeItemReference) (*dto.ItemSettings, error)
	SaveSettings(ctx context.Context, request dto.SaveItemSettings) (*dto.ItemSettings, error)
}
