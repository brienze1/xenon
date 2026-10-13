package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func (t *tools) gitPush(ctx context.Context, call Call) (Result, error) {
	status, err := t.controllers.SyncTransfer.Push(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status)}, nil
}
