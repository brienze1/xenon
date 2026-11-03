package usecase

import "context"

type LockWorkspace interface {
	// LockWorkspace is reentrant for the returned context: a nested call with it for the same workspace does not block.
	LockWorkspace(ctx context.Context, localId string) (context.Context, func())
}
