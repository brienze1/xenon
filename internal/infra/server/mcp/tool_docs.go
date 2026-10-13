package mcp

import "context"

func (t *tools) docs(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
