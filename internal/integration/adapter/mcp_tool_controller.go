package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type MCPToolController interface {
	Tool() dto.MCPTool
	Handle(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error)
}
