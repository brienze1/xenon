package mcp

import "context"

func (t *tools) gitPush(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
