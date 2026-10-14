package mcp

import (
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func newToolDefinition(tool dto.MCPTool) *sdk.Tool {
	destructive := tool.Destructive
	openWorld := tool.OpenWorld
	annotations := &sdk.ToolAnnotations{
		ReadOnlyHint:   tool.ReadOnly,
		IdempotentHint: tool.Idempotent,
		OpenWorldHint:  &openWorld,
	}
	if !tool.ReadOnly {
		annotations.DestructiveHint = &destructive
	}
	return &sdk.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: tool.InputSchema(),
		Annotations: annotations,
	}
}
