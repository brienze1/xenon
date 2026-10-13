package mcp

import "context"

func (t *tools) workspaceEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
