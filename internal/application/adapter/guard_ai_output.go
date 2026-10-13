package adapter

import "context"

type GuardAIOutputService interface {
	Guard(ctx context.Context, localId, text string) string
	Register(ctx context.Context, values ...string)
}
