package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpWorkspaceCreateAction = "create"
	mcpWorkspaceOpenAction   = "open"
	mcpWorkspaceSwitchAction = "switch"
	mcpWorkspaceRenameAction = "rename"
)

var workspaceEditAITool = entity.AITool{
	Name:        "workspace_edit",
	Description: "Create, open, switch or rename a workspace.",
	Actions:     []string{mcpWorkspaceCreateAction, mcpWorkspaceOpenAction, mcpWorkspaceSwitchAction, mcpWorkspaceRenameAction},
	Effect:      aitooleffect.Write,
}

type mcpWorkspaceEdit struct {
	errorHandler             adapter.ErrorHandler
	createWorkspaceService   appAdapter.CreateWorkspaceService
	openWorkspaceService     appAdapter.OpenWorkspaceService
	renameWorkspaceService   appAdapter.RenameWorkspaceService
	activateWorkspaceService appAdapter.ActivateWorkspaceService
	autoSyncService          appAdapter.AutoSyncService
}

func NewMCPWorkspaceEdit(
	errorHandler adapter.ErrorHandler,
	createWorkspaceService appAdapter.CreateWorkspaceService,
	openWorkspaceService appAdapter.OpenWorkspaceService,
	renameWorkspaceService appAdapter.RenameWorkspaceService,
	activateWorkspaceService appAdapter.ActivateWorkspaceService,
	autoSyncService appAdapter.AutoSyncService,
) adapter.MCPToolController {
	return &mcpWorkspaceEdit{
		errorHandler:             errorHandler,
		createWorkspaceService:   createWorkspaceService,
		openWorkspaceService:     openWorkspaceService,
		renameWorkspaceService:   renameWorkspaceService,
		activateWorkspaceService: activateWorkspaceService,
		autoSyncService:          autoSyncService,
	}
}

func (m *mcpWorkspaceEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(workspaceEditAITool)
}

func (m *mcpWorkspaceEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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

	var arguments dto.MCPWorkspaceEdit
	err = call.Decode(&arguments)
	if err != nil {
		return result, err
	}

	switch call.Action {
	case mcpWorkspaceCreateAction:
		result, err = m.create(ctx, arguments)
	case mcpWorkspaceOpenAction:
		result, err = m.open(ctx, arguments)
	case mcpWorkspaceSwitchAction:
		result, err = m.switchTo(ctx, arguments)
	case mcpWorkspaceRenameAction:
		result, err = m.rename(ctx, arguments)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpWorkspaceEdit) create(ctx context.Context, arguments dto.MCPWorkspaceEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateWorkspace()
	err := validateCreateWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	workspace, err := m.createWorkspaceService.Create(ctx, request.Name, request.GetPath())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return editedWorkspace(*workspace, aichangekind.Workspaces), nil
}

func (m *mcpWorkspaceEdit) open(ctx context.Context, arguments dto.MCPWorkspaceEdit) (dto.MCPToolResult, error) {
	request := arguments.OpenWorkspace()
	err := validateOpenWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	workspace, err := m.openWorkspaceService.Open(ctx, request.GetPath())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return editedWorkspace(*workspace, aichangekind.Workspaces), nil
}

func (m *mcpWorkspaceEdit) switchTo(ctx context.Context, arguments dto.MCPWorkspaceEdit) (dto.MCPToolResult, error) {
	request := arguments.ActivateWorkspace()
	err := validateActivateWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	snapshot, err := m.activateWorkspaceService.Activate(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Opened, request.LocalId)
	return editedWorkspace(snapshot.Workspace, aichangekind.Workspaces, aichangekind.Tree), nil
}

func (m *mcpWorkspaceEdit) rename(ctx context.Context, arguments dto.MCPWorkspaceEdit) (dto.MCPToolResult, error) {
	request := arguments.RenameWorkspace()
	err := validateRenameWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	workspace, err := m.renameWorkspaceService.Rename(ctx, request.LocalId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return editedWorkspace(*workspace, aichangekind.Workspaces), nil
}

func editedWorkspace(workspace entity.RegisteredWorkspace, changes ...enums.AIChangeKind) dto.MCPToolResult {
	return dto.NewMCPToolResult(dto.NewMCPWorkspace(workspace)).WithTarget(workspace.Target()).WithChanges(changes...)
}
