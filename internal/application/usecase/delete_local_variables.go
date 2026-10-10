package usecase

import "context"

type DeleteLocalVariables interface {
	Delete(ctx context.Context, localId string) error
}
