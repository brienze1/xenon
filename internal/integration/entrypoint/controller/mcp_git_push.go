package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

var gitPushAITool = entity.AITool{
	Name:            "git_push",
	Description:     "Push committed changes to the linked remote.",
	Effect:          aitooleffect.Destructive,
	OpenWorld:       true,
	WorkspaceScoped: true,
}

type mcpGitPush struct {
	errorHandler         adapter.ErrorHandler
	pushWorkspaceService appAdapter.PushWorkspaceService
}

func NewMCPGitPush(errorHandler adapter.ErrorHandler, pushWorkspaceService appAdapter.PushWorkspaceService) adapter.MCPToolController {
	return &mcpGitPush{errorHandler: errorHandler, pushWorkspaceService: pushWorkspaceService}
}

func (m *mcpGitPush) Tool() dto.MCPTool {
	return dto.NewMCPTool(gitPushAITool)
}

func (m *mcpGitPush) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
	ctx := logger.GetContext(toolCtx)
	logger.Info(ctx, "Started", call)

	defer func(ctx context.Context, call any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, call, recovered)
		}
		if err != nil {
			err = m.errorHandler.Handle(ctx, err, call)
		}
	}(ctx, call)

	request := dto.NewSyncWorkspaceReference(call.LocalId)
	err = validatePushWorkspace(&request)
	if err != nil {
		return result, err
	}

	status, err := m.pushWorkspaceService.Push(ctx, request.LocalId)
	if err != nil {
		return result, err
	}
	result = dto.NewMCPToolResult(dto.NewMCPGitStatus(*status))

	logger.Info(ctx, "Finished", result)
	return result, nil
}
