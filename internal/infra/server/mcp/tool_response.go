package mcp

import "context"

func (t *tools) response(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
