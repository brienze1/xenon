package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type MCPCallController interface {
	Call(ctx context.Context, call dto.MCPCall, tool MCPToolController) dto.MCPOutcome
	CloseSession(ctx context.Context, session dto.MCPSession) error
}
