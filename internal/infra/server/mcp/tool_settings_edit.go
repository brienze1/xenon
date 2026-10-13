package mcp

import "context"

func (t *tools) settingsEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
