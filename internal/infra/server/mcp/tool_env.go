package mcp

import "context"

func (t *tools) env(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
