package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func validateMCPRunTarget(call dto.MCPToolCall, target dto.MCPRunTarget) error {
	if target.IsAmbiguous() {
		return mcpErrors.RejectInvalidArguments(call.Tool, mcpRunTargetConflict)
	}
	if target.IsMissing() {
		return mcpErrors.RejectInvalidArguments(call.Tool, mcpRunTargetMissing)
	}
	return nil
}
