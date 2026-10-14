package adapter

import "context"

type GuardAIOutputService interface {
	// Guard always returns text masked with the secrets it could collect, also alongside an error.
	Guard(ctx context.Context, localId, text string) (string, error)
}
