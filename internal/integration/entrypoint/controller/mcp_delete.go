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
	mcpDeleteSelector           = "target"
	mcpDeleteItemTarget         = "item"
	mcpDeleteEnvironmentTarget  = "environment"
	mcpDeleteCookieTarget       = "cookie"
	mcpDeleteCookieDomainTarget = "cookie_domain"
	mcpDeleteCookiesTarget      = "cookies"
	mcpDeleteWorkspaceTarget    = "workspace"
	mcpDeleteTokenTarget        = "token"
	mcpDeleteGitLinkTarget      = "git_link"
	mcpDeleteHistoryTarget      = "history"
	mcpDeleteConsoleTarget      = "console"
	mcpDeleteAliasTarget        = "alias"
	mcpDeleteAliasReason        = "owner_id must be the request the alias belongs to"
	mcpUnlinkPathReason         = "path (an empty folder that keeps the workspace files) is required to unlink"
)

var deleteAITool = entity.AITool{
	Name:        "delete",
	Description: "Delete an item, a request alias (id + owner_id), environment, cookie(s), workspace, OAuth token, git link, History or the console.",
	Selector:    mcpDeleteSelector,
	Actions: []string{
		mcpDeleteItemTarget,
		mcpDeleteAliasTarget,
		mcpDeleteEnvironmentTarget,
		mcpDeleteCookieTarget,
		mcpDeleteCookieDomainTarget,
		mcpDeleteCookiesTarget,
		mcpDeleteWorkspaceTarget,
		mcpDeleteTokenTarget,
		mcpDeleteGitLinkTarget,
		mcpDeleteHistoryTarget,
		mcpDeleteConsoleTarget,
	},
	Effect:          aitooleffect.Destructive,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpDelete struct {
	errorHandler             adapter.ErrorHandler
	searchTreeService        appAdapter.SearchTreeService
	manageTreeItemService    appAdapter.ManageTreeItemService
	manageExampleService     appAdapter.ManageExampleService
	manageAliasService       appAdapter.ManageAliasService
	manageEnvironmentService appAdapter.ManageEnvironmentService
	manageCookiesService     appAdapter.ManageCookiesService
	removeWorkspaceService   appAdapter.RemoveWorkspaceService
	manageOAuthService       appAdapter.ManageOAuthService
	unlinkWorkspaceService   appAdapter.UnlinkWorkspaceService
	manageHistoryService     appAdapter.ManageHistoryService
	manageConsoleService     appAdapter.ManageConsoleService
}

func NewMCPDelete(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	manageTreeItemService appAdapter.ManageTreeItemService,
	manageExampleService appAdapter.ManageExampleService,
	manageAliasService appAdapter.ManageAliasService,
	manageEnvironmentService appAdapter.ManageEnvironmentService,
	manageCookiesService appAdapter.ManageCookiesService,
	removeWorkspaceService appAdapter.RemoveWorkspaceService,
	manageOAuthService appAdapter.ManageOAuthService,
	unlinkWorkspaceService appAdapter.UnlinkWorkspaceService,
	manageHistoryService appAdapter.ManageHistoryService,
	manageConsoleService appAdapter.ManageConsoleService,
) adapter.MCPToolController {
	return &mcpDelete{
		errorHandler:             errorHandler,
		searchTreeService:        searchTreeService,
		manageTreeItemService:    manageTreeItemService,
		manageExampleService:     manageExampleService,
		manageAliasService:       manageAliasService,
		manageEnvironmentService: manageEnvironmentService,
		manageCookiesService:     manageCookiesService,
		removeWorkspaceService:   removeWorkspaceService,
		manageOAuthService:       manageOAuthService,
		unlinkWorkspaceService:   unlinkWorkspaceService,
		manageHistoryService:     manageHistoryService,
		manageConsoleService:     manageConsoleService,
	}
}

func (m *mcpDelete) Tool() dto.MCPTool {
	return dto.NewMCPTool(deleteAITool, dto.NewMCPToolArguments[dto.MCPDelete](deleteAITool.Actions...))
}

func (m *mcpDelete) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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

	var arguments dto.MCPDelete
	err = call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	switch call.Action {
	case mcpDeleteItemTarget:
		result, err = m.item(ctx, call, arguments)
	case mcpDeleteAliasTarget:
		result, err = m.alias(ctx, call, arguments)
	case mcpDeleteEnvironmentTarget:
		result, err = m.environment(ctx, call, arguments)
	case mcpDeleteCookieTarget:
		result, err = m.cookie(ctx, call, arguments)
	case mcpDeleteCookieDomainTarget:
		result, err = m.cookieDomain(ctx, call, arguments)
	case mcpDeleteCookiesTarget:
		result, err = m.cookies(ctx, call)
	case mcpDeleteWorkspaceTarget:
		result, err = m.workspace(ctx, call, arguments)
	case mcpDeleteTokenTarget:
		result, err = m.token(ctx, call, arguments)
	case mcpDeleteGitLinkTarget:
		result, err = m.gitLink(ctx, call, arguments)
	case mcpDeleteHistoryTarget:
		result, err = m.history(ctx, call, arguments)
	case mcpDeleteConsoleTarget:
		result, err = m.console(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpDelete) item(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.ItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	if entry.IsExample() {
		err = m.deleteExample(ctx, call, entry)
	} else {
		err = m.deleteTreeItem(ctx, call, entry)
	}
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, entry.Target(), aichangekind.Tree), nil
}

func (m *mcpDelete) alias(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.OwnerItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !entry.IsRequest() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpDeleteAliasReason)
	}

	request := arguments.AliasReference(call.LocalId, entry.ID())
	err = validateDeleteAlias(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if _, err = m.manageAliasService.Delete(ctx, request.LocalId, request.RequestId, request.AliasId, request.BaseRevision, true); err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, entry.Target(), aichangekind.Tree), nil
}

func (m *mcpDelete) deleteExample(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) error {
	request := dto.NewExampleReference(call.LocalId, entry.RequestID, entry.ID())
	err := validateDeleteExample(&request)
	if err != nil {
		return err
	}
	return m.manageExampleService.Delete(ctx, request.LocalId, request.RequestId, request.ExampleId)
}

func (m *mcpDelete) deleteTreeItem(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) error {
	request := dto.NewTreeItemReference(call.LocalId, entry.ID())
	err := validateDeleteItem(&request)
	if err != nil {
		return err
	}
	_, err = m.manageTreeItemService.Delete(ctx, request.LocalId, request.Id)
	return err
}

func (m *mcpDelete) environment(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.EnvironmentReference(call.LocalId)
	err := validateDeleteEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	_, err = m.manageEnvironmentService.Delete(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, arguments.IdTarget(), aichangekind.Variables), nil
}

func (m *mcpDelete) cookie(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.DeleteCookie(call.LocalId)
	err := validateDeleteCookie(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	cookie := request.ToEntity()
	err = m.manageCookiesService.Delete(ctx, request.LocalId, cookie)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, cookie.ActivityTarget(), aichangekind.Cookies), nil
}

func (m *mcpDelete) cookieDomain(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.DeleteCookieDomain(call.LocalId)
	err := validateDeleteCookieDomain(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = m.manageCookiesService.DeleteDomain(ctx, request.LocalId, request.Domain)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, arguments.DomainTarget(), aichangekind.Cookies), nil
}

func (m *mcpDelete) cookies(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	request := dto.NewCookieJarReference(call.LocalId)
	err := validateClearCookies(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = m.manageCookiesService.Clear(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, entity.AIActivityTarget{}, aichangekind.Cookies), nil
}

func (m *mcpDelete) workspace(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.RemoveWorkspace()
	err := validateRemoveWorkspace(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	_, err = m.removeWorkspaceService.Remove(ctx, request.LocalId, request.MoveToTrash)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, arguments.IdTarget(), aichangekind.Workspaces), nil
}

func (m *mcpDelete) token(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.OAuthTokenReference(call.LocalId)
	err := validateClearOAuthToken(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = m.manageOAuthService.Clear(ctx, request.LocalId, request.OwnerId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, arguments.OwnerTarget()), nil
}

func (m *mcpDelete) gitLink(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	if !arguments.HasPath() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpUnlinkPathReason)
	}

	request := arguments.UnlinkRepository(call.LocalId)
	err := validateUnlinkRepository(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	registered, err := m.unlinkWorkspaceService.Unlink(ctx, request.LocalId, request.GetTargetFolder())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, registered.Target(), aichangekind.Workspaces), nil
}

func (m *mcpDelete) history(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPDelete) (dto.MCPToolResult, error) {
	request := arguments.ClearHistory(call.LocalId)
	err := validateClearHistory(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = m.manageHistoryService.Clear(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return newMCPDeletedResult(call, arguments.SourceTarget(), aichangekind.History), nil
}

func (m *mcpDelete) console(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	m.manageConsoleService.Clear(ctx)
	return newMCPDeletedResult(call, entity.AIActivityTarget{}), nil
}

func newMCPDeletedResult(call dto.MCPToolCall, target entity.AIActivityTarget, kinds ...enums.AIChangeKind) dto.MCPToolResult {
	return dto.NewMCPToolResult(dto.NewMCPDeleted(call.Action, target)).WithTarget(target).WithChanges(kinds...)
}
