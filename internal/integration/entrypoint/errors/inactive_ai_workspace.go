package errors

import "github.com/brienze1/xenon/pkg/errs"

func RejectInactiveAIWorkspace() error {
	return errs.NotFoundError("No active workspace — create or open one", "MCP-06404")
}
