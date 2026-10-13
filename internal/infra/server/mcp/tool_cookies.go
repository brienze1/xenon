package mcp

import "context"

func (t *tools) cookies(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
