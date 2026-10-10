package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

// SendScriptRequest sends what scripts ask for: pm.sendRequest (resolved with the scopes as the script left them) and pm.execution.runRequest (a saved request with its own scripts).
type SendScriptRequest interface {
	Send(ctx context.Context, request entity.Request, scopes entity.ScriptScopes) entity.NestedRequest
	Run(ctx context.Context, requestId string, scopes entity.ScriptScopes) entity.NestedRequest
}
