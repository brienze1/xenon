package usecase

import "context"

type ClearConsole interface {
	Clear(ctx context.Context)
}
