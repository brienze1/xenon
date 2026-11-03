package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAISessionService interface {
	Open(ctx context.Context, sessionId string) (*entity.AISessionContext, error)
	Current(ctx context.Context, sessionId string) (*entity.AISessionContext, error)
	// Enter makes localId the session's workspace and returns that workspace.
	Enter(ctx context.Context, sessionId, localId string) (*entity.RegisteredWorkspace, error)
	ChooseEnvironment(ctx context.Context, sessionId, environmentId string) (*entity.AISessionContext, error)
	ResolveEnvironment(ctx context.Context, localId, ref string) (string, bool, error)
	Close(ctx context.Context, sessionId string) error
}
