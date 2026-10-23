package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func resolveMCPRunTarget(ctx context.Context, searchTreeService appAdapter.SearchTreeService, call dto.MCPToolCall, target dto.MCPRunTarget) (string, entity.AIActivityTarget, error) {
	if target.IsAmbiguous() {
		return "", entity.AIActivityTarget{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpRunTargetConflict)
	}
	if target.IsMissing() {
		return "", entity.AIActivityTarget{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpRunTargetMissing)
	}

	tree, err := loadMCPTree(ctx, searchTreeService, call.LocalId)
	if err != nil {
		return "", entity.AIActivityTarget{}, err
	}
	if target.IsPath() {
		if entry, found := tree.ContainerAt(target.Path()); found {
			return entry.ID(), entry.Target(), nil
		}
		return target.Path(), entity.NewAIItemTarget(target.Path()), nil
	}
	if entry, found := tree.Entry(target.TargetItemId()); found {
		return entry.ID(), entry.Target(), nil
	}
	return target.TargetItemId(), entity.NewAIItemTarget(target.TargetItemId()), nil
}
