package usecase

import "context"

type DeleteUIState interface {
	Delete(ctx context.Context, localId string) error
}
