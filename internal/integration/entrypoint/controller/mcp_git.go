package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpGitStatusAction  = "status"
	mcpGitChangesAction = "changes"
	mcpGitSessionAction = "session"
)

var gitAITool = entity.AITool{
	Name:            "git",
	Description:     "Show git sync status and pending changes.",
	Actions:         []string{mcpGitStatusAction, mcpGitChangesAction, mcpGitSessionAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpGit struct {
	errorHandler            adapter.ErrorHandler
	manageSyncService       appAdapter.ManageSyncService
	resolveConflictsService appAdapter.ResolveConflictsService
}

func NewMCPGit(
	errorHandler adapter.ErrorHandler,
	manageSyncService appAdapter.ManageSyncService,
	resolveConflictsService appAdapter.ResolveConflictsService,
) adapter.MCPToolController {
	return &mcpGit{errorHandler: errorHandler, manageSyncService: manageSyncService, resolveConflictsService: resolveConflictsService}
}

func (m *mcpGit) Tool() dto.MCPTool {
	return dto.NewMCPTool(gitAITool)
}

func (m *mcpGit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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

	switch call.Action {
	case mcpGitStatusAction:
		result, err = m.status(ctx, call)
	case mcpGitChangesAction:
		result, err = m.changes(ctx, call)
	case mcpGitSessionAction:
		result, err = m.session(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpGit) status(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	status, err := loadMCPSyncStatus(ctx, m.manageSyncService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(status)), nil
}

func (m *mcpGit) changes(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	changes, err := loadMCPSyncChanges(ctx, m.manageSyncService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitChangeList(changes)), nil
}

func (m *mcpGit) session(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	session, err := loadMCPSyncSession(ctx, m.resolveConflictsService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitSession(session)), nil
}
