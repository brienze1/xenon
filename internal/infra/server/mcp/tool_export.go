package mcp

import "context"

func (t *tools) export(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
