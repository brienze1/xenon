package mcp

import "context"

func (t *tools) git(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
