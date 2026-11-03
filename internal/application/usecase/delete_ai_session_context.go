package usecase

import "context"

type DeleteAISessionContext interface {
	Delete(ctx context.Context, sessionId string)
}
