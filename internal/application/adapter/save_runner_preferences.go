package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRunnerPreferencesService interface {
	Save(ctx context.Context, localId, itemId string, preferences entity.RunnerPreferences) error
}
