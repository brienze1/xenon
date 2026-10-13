package mcp

import "context"

func (t *tools) envEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
