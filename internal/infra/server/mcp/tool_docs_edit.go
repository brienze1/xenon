package mcp

import "context"

func (t *tools) docsEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
