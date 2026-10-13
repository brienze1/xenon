package mcp

import "context"

func (t *tools) workspace(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
