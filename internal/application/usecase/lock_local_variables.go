package usecase

import "context"

type LockLocalVariables interface {
	LockVariables(ctx context.Context, localId string) (unlock func())
}
