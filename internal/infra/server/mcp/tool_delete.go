package mcp

import "context"

func (t *tools) delete(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
