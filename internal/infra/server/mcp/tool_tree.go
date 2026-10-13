package mcp

import "context"

func (t *tools) tree(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
