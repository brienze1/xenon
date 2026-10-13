package mcp

import "context"

func (t *tools) history(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
