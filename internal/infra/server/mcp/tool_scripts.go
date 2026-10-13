package mcp

import "context"

func (t *tools) scripts(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
