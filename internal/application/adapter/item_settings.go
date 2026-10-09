package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ItemSettingsService interface {
	Get(ctx context.Context, localId, id string) (*entity.Collection, error)
	Save(ctx context.Context, localId string, settings entity.Collection) (*entity.Collection, error)
}
