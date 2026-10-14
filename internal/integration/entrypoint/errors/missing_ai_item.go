package errors

import "github.com/brienze1/xenon/pkg/errs"

func RejectMissingAIItem(id string) error {
	return errs.NotFoundError("No item with id \""+id+"\" in the active workspace", "MCP-08404")
}
