package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindRunnerPreferences interface {
	FindRunnerPreferences(ctx context.Context, localId string) (map[string]entity.RunnerPreferences, error)
}
