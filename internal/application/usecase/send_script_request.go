package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

// SendScriptRequest sends a request a script asked for (pm.sendRequest), resolved with the send's local variables.
type SendScriptRequest interface {
	Send(ctx context.Context, request entity.Request, localVariables map[string]any) entity.NestedRequest
}
