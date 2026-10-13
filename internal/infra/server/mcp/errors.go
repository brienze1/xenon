package mcp

import (
	"strings"

	"github.com/brienze1/xenon/pkg/errs"
)

func invalidArguments(tool, reason string) error {
	return errs.BadRequestError("Invalid arguments for "+tool+": "+reason, "MCP-07400")
}

func unknownAction(tool, action string, actions []string) error {
	return errs.BadRequestError("Unknown action \""+action+"\" for "+tool+" — use one of: "+strings.Join(actions, ", "), "MCP-07400",
		errs.ErrorDetails{Attribute: "action", Messages: actions})
}

func noActiveWorkspace() error {
	return errs.NotFoundError("No active workspace — create or open one", "MCP-06404")
}

func humanOnly(what string) error {
	return errs.ForbiddenError(what+" is human-only — do it in Xenon.", "MCP-02403")
}

func itemNotFound(id string) error {
	return errs.NotFoundError("No item with id \""+id+"\" in the active workspace", "MCP-08404")
}

func unsupportedAction(tool, action string) error {
	return errs.BadRequestError("The action \""+action+"\" of "+tool+" is not available", "MCP-07400")
}
