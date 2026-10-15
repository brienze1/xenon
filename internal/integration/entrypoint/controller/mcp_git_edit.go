package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpGitEditLinkAction       = "link"
	mcpGitEditCloneAction      = "clone"
	mcpGitEditSettingsAction   = "settings"
	mcpGitEditCommitAction     = "commit"
	mcpGitEditPullAction       = "pull"
	mcpGitEditResolveAction    = "resolve"
	mcpGitEditFinishPullAction = "finish_pull"
	mcpGitEditCancelPullAction = "cancel_pull"
	mcpNoOpenPullReason        = "there is no open pull to "
)

var gitEditAITool = entity.AITool{
	Name:        "git_edit",
	Description: "Link, configure, commit, pull and resolve git sync for the workspace.",
	Actions: []string{
		mcpGitEditLinkAction,
		mcpGitEditCloneAction,
		mcpGitEditSettingsAction,
		mcpGitEditCommitAction,
		mcpGitEditPullAction,
		mcpGitEditResolveAction,
		mcpGitEditFinishPullAction,
		mcpGitEditCancelPullAction,
	},
	Effect:          aitooleffect.Destructive,
	OpenWorld:       true,
	WorkspaceScoped: true,
}

type mcpGitEdit struct {
	errorHandler            adapter.ErrorHandler
	manageSyncService       appAdapter.ManageSyncService
	linkWorkspaceService    appAdapter.LinkWorkspaceService
	cloneWorkspaceService   appAdapter.CloneWorkspaceService
	commitChangesService    appAdapter.CommitChangesService
	pullWorkspaceService    appAdapter.PullWorkspaceService
	resolveConflictsService appAdapter.ResolveConflictsService
	finishPullService       appAdapter.FinishPullService
}

func NewMCPGitEdit(
	errorHandler adapter.ErrorHandler,
	manageSyncService appAdapter.ManageSyncService,
	linkWorkspaceService appAdapter.LinkWorkspaceService,
	cloneWorkspaceService appAdapter.CloneWorkspaceService,
	commitChangesService appAdapter.CommitChangesService,
	pullWorkspaceService appAdapter.PullWorkspaceService,
	resolveConflictsService appAdapter.ResolveConflictsService,
	finishPullService appAdapter.FinishPullService,
) adapter.MCPToolController {
	return &mcpGitEdit{
		errorHandler:            errorHandler,
		manageSyncService:       manageSyncService,
		linkWorkspaceService:    linkWorkspaceService,
		cloneWorkspaceService:   cloneWorkspaceService,
		commitChangesService:    commitChangesService,
		pullWorkspaceService:    pullWorkspaceService,
		resolveConflictsService: resolveConflictsService,
		finishPullService:       finishPullService,
	}
}

func (m *mcpGitEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(
		gitEditAITool,
		dto.NewMCPToolArguments[dto.MCPGitLink](mcpGitEditLinkAction, mcpGitEditCloneAction),
		dto.NewMCPToolArguments[dto.MCPGitSettings](mcpGitEditSettingsAction),
		dto.NewMCPToolArguments[dto.MCPGitCommit](mcpGitEditCommitAction),
		dto.NewMCPToolArguments[dto.MCPGitResolve](mcpGitEditResolveAction),
	)
}

func (m *mcpGitEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpGitEditLinkAction:
		result, err = m.link(ctx, call)
	case mcpGitEditCloneAction:
		result, err = m.clone(ctx, call)
	case mcpGitEditSettingsAction:
		result, err = m.settings(ctx, call)
	case mcpGitEditCommitAction:
		result, err = m.commit(ctx, call)
	case mcpGitEditPullAction:
		result, err = m.pull(ctx, call)
	case mcpGitEditResolveAction:
		result, err = m.resolve(ctx, call)
	case mcpGitEditFinishPullAction:
		result, err = m.finishPull(ctx, call)
	case mcpGitEditCancelPullAction:
		result, err = m.cancelPull(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpGitEdit) link(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPGitLink
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.LinkRepository(call.LocalId)
	err = validateLinkRepository(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	status, err := m.linkWorkspaceService.Link(ctx, request.LocalId, request.ToEntity(), request.AccountID, request.ClonePath, request.LinkMode())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(*status)).WithChanges(aichangekind.Tree), nil
}

func (m *mcpGitEdit) clone(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPGitLink
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.CloneWorkspace()
	err = validateCloneWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	registered, err := m.cloneWorkspaceService.Clone(ctx, request.ToEntity(), request.AccountID, request.ClonePath, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewWorkspaceSummary(*registered)).WithTarget(registered.NamedTarget()).WithChanges(aichangekind.Workspaces), nil
}

func (m *mcpGitEdit) settings(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPGitSettings
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	current, err := loadMCPSyncStatus(ctx, m.manageSyncService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.SaveSyncSettings(call.LocalId, current.Settings)
	err = validateSaveSyncSettings(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = m.manageSyncService.SaveSettings(ctx, request.LocalId, request.AccountID, request.AutoPull, request.AutoPush)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	saved, err := loadMCPSyncStatus(ctx, m.manageSyncService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(saved)), nil
}

func (m *mcpGitEdit) commit(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPGitCommit
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	var changedIds []string
	if !arguments.HasItemIds() {
		changes, err := loadMCPSyncChanges(ctx, m.manageSyncService, call.LocalId)
		if err != nil {
			return dto.MCPToolResult{}, err
		}
		changedIds = entity.ChangeItemIDs(changes)
	}

	request := arguments.CommitSyncChanges(call.LocalId, changedIds)
	err = validateCommitSyncChanges(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	status, err := m.commitChangesService.Commit(ctx, request.LocalId, request.Message, request.ItemIDs)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(*status)), nil
}

func (m *mcpGitEdit) pull(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	request := dto.NewSyncWorkspaceReference(call.LocalId)
	err := validatePullWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	pulled, err := m.pullWorkspaceService.Pull(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitPull(*pulled)).WithChanges(aichangekind.Tree), nil
}

func (m *mcpGitEdit) resolve(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPGitResolve
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	sessionId, err := m.openPullId(ctx, call)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	for _, request := range arguments.ResolveSyncConflicts(sessionId) {
		err = validateResolveSyncConflict(&request)
		if err != nil {
			return dto.MCPToolResult{}, err
		}

		_, err = m.resolveConflictsService.Resolve(ctx, request.SessionId, request.ItemId, request.ConflictResolution(), request.FieldPath)
		if err != nil {
			return dto.MCPToolResult{}, err
		}
	}

	session, err := loadMCPSyncSession(ctx, m.resolveConflictsService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitSession(session)), nil
}

func (m *mcpGitEdit) finishPull(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	sessionId, err := m.openPullId(ctx, call)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := dto.NewSyncSessionReference(sessionId)
	err = validateFinishPull(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	status, err := m.finishPullService.Finish(ctx, request.SessionId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(*status)).WithChanges(aichangekind.Tree), nil
}

func (m *mcpGitEdit) cancelPull(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	sessionId, err := m.openPullId(ctx, call)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := dto.NewSyncSessionReference(sessionId)
	err = validateCancelPull(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	status, err := m.resolveConflictsService.Cancel(ctx, request.SessionId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPGitStatus(*status)).WithChanges(aichangekind.Tree), nil
}

func (m *mcpGitEdit) openPullId(ctx context.Context, call dto.MCPToolCall) (string, error) {
	session, err := loadMCPSyncSession(ctx, m.resolveConflictsService, call.LocalId)
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", mcpErrors.RejectInvalidArguments(call.Tool, mcpNoOpenPullReason+call.Action)
	}
	return session.ID, nil
}
