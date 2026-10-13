package mcp

import "context"

func (t *tools) gitEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
