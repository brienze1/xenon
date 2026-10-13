package mcp

import "context"

func (t *tools) treeEdit(ctx context.Context, call Call) (Result, error) {
	return Result{}, call.Unsupported()
}
