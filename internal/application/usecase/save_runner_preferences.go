package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRunnerPreferences interface {
	SaveRunnerPreferences(ctx context.Context, localId, itemId string, preferences entity.RunnerPreferences) error
	DeleteRunnerPreferences(ctx context.Context, localId string, itemIds []string) error
}
