package usecase

import "context"

type DeleteRunnerPreferences interface {
	DeleteRunnerPreferences(ctx context.Context, localId string, itemIds []string) error
}
