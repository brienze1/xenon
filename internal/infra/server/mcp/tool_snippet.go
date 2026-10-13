package mcp

import "context"

func (t *tools) snippet(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
