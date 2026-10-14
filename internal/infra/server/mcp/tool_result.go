package mcp

import (
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func newToolResult(outcome dto.MCPOutcome) *sdk.CallToolResult {
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: outcome.Text}}, IsError: outcome.Failed}
}
