package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpTreeEditCreateAction       = "create"
	mcpTreeEditUpdateAction       = "update"
	mcpTreeEditRenameAction       = "rename"
	mcpTreeEditMoveAction         = "move"
	mcpTreeEditDuplicateAction    = "duplicate"
	mcpTreeEditSaveSettingsAction = "save_settings"
	mcpTreeEditGetTokenAction     = "get_token"
	mcpTreeEditRefreshTokenAction = "refresh_token"
	mcpTreeEditUseTokenAction     = "use_token"

	mcpTreeEditKindReason      = "kind must be collection, folder, request, example or alias"
	mcpTreeEditMoveReason      = "an example cannot be moved"
	mcpTreeEditUpdateReason    = "update edits a request or an example; use rename or save_settings for a collection or folder"
	mcpTreeEditVariablesReason = "only a collection has variables"
	mcpTreeEditRequestReason   = "request_id must be a request"
)

var treeEditAITool = entity.AITool{
	Name:        "tree_edit",
	Description: "Create, update, rename, move or duplicate items, create, update or rename request aliases (kind alias with request_id), edit collection/folder settings, and manage OAuth tokens.",
	Actions: []string{
		mcpTreeEditCreateAction, mcpTreeEditUpdateAction, mcpTreeEditRenameAction, mcpTreeEditMoveAction, mcpTreeEditDuplicateAction,
		mcpTreeEditSaveSettingsAction, mcpTreeEditGetTokenAction, mcpTreeEditRefreshTokenAction, mcpTreeEditUseTokenAction,
	},
	Effect:          aitooleffect.Destructive,
	OpenWorld:       true,
	WorkspaceScoped: true,
}

type mcpTreeEdit struct {
	errorHandler               adapter.ErrorHandler
	searchTreeService          appAdapter.SearchTreeService
	manageTreeItemService      appAdapter.ManageTreeItemService
	saveRequestService         appAdapter.SaveRequestService
	autoSyncService            appAdapter.AutoSyncService
	manageExampleService       appAdapter.ManageExampleService
	manageAliasService         appAdapter.ManageAliasService
	manageResponseService      appAdapter.ManageResponseService
	guardAIOutputService       appAdapter.GuardAIOutputService
	manageItemSettingsService  appAdapter.ManageItemSettingsService
	manageVariableTableService appAdapter.ManageVariableTableService
	manageOAuthService         appAdapter.ManageOAuthService
}

func NewMCPTreeEdit(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	manageTreeItemService appAdapter.ManageTreeItemService,
	saveRequestService appAdapter.SaveRequestService,
	autoSyncService appAdapter.AutoSyncService,
	manageExampleService appAdapter.ManageExampleService,
	manageAliasService appAdapter.ManageAliasService,
	manageResponseService appAdapter.ManageResponseService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	manageItemSettingsService appAdapter.ManageItemSettingsService,
	manageVariableTableService appAdapter.ManageVariableTableService,
	manageOAuthService appAdapter.ManageOAuthService,
) adapter.MCPToolController {
	return &mcpTreeEdit{
		errorHandler:               errorHandler,
		searchTreeService:          searchTreeService,
		manageTreeItemService:      manageTreeItemService,
		saveRequestService:         saveRequestService,
		autoSyncService:            autoSyncService,
		manageExampleService:       manageExampleService,
		manageAliasService:         manageAliasService,
		manageResponseService:      manageResponseService,
		guardAIOutputService:       guardAIOutputService,
		manageItemSettingsService:  manageItemSettingsService,
		manageVariableTableService: manageVariableTableService,
		manageOAuthService:         manageOAuthService,
	}
}

func (m *mcpTreeEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(treeEditAITool, dto.NewMCPToolArguments[dto.MCPTreeEdit](treeEditAITool.Actions...))
}

func (m *mcpTreeEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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

	var arguments dto.MCPTreeEdit
	err = call.Decode(&arguments)
	if err != nil {
		return result, err
	}

	result, err = m.edit(ctx, call, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	result = result.Prepended(aichangekind.Tree)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpTreeEdit) edit(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	if arguments.EditsAlias() {
		return m.editAlias(ctx, call, arguments)
	}
	switch call.Action {
	case mcpTreeEditCreateAction:
		return m.create(ctx, call, arguments)
	case mcpTreeEditGetTokenAction, mcpTreeEditRefreshTokenAction, mcpTreeEditUseTokenAction:
		return m.token(ctx, call, arguments)
	}

	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.ItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	switch call.Action {
	case mcpTreeEditUpdateAction:
		return m.update(ctx, call, entry, arguments)
	case mcpTreeEditRenameAction:
		return m.rename(ctx, call, entry, arguments)
	case mcpTreeEditMoveAction:
		return m.move(ctx, call, entry, arguments)
	case mcpTreeEditDuplicateAction:
		return m.duplicate(ctx, call, entry)
	case mcpTreeEditSaveSettingsAction:
		return m.saveSettings(ctx, call, entry, arguments)
	default:
		return dto.MCPToolResult{}, mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
}

func (m *mcpTreeEdit) create(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	switch arguments.ItemKind() {
	case itemkind.Collection:
		return m.createCollection(ctx, call, arguments)
	case itemkind.Folder:
		return m.createFolder(ctx, call, arguments)
	case itemkind.Request:
		return m.createRequest(ctx, call, arguments)
	case itemkind.Example:
		return m.createExample(ctx, call, arguments)
	default:
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditKindReason)
	}
}

func (m *mcpTreeEdit) createCollection(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateCollection(call.LocalId)
	err := validateCreateCollection(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	node, err := m.manageTreeItemService.CreateCollection(ctx, request.LocalId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewTreeNode(*node).Id, "")
}

func (m *mcpTreeEdit) createFolder(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateFolder(call.LocalId)
	err := validateCreateFolder(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	node, err := m.manageTreeItemService.CreateFolder(ctx, request.LocalId, request.ParentId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewTreeNode(*node).Id, "")
}

func (m *mcpTreeEdit) rename(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	if entry.IsExample() {
		return m.renameExample(ctx, call, entry, arguments)
	}

	request := arguments.RenameItem(call.LocalId, entry.ID())
	err := validateRenameItem(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	node, err := m.manageTreeItemService.Rename(ctx, request.LocalId, request.Id, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewTreeNode(*node).Id, "")
}

func (m *mcpTreeEdit) move(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	if entry.IsExample() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditMoveReason)
	}

	request := arguments.MoveItem(call.LocalId, entry.ID())
	err := validateMoveItem(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	_, err = m.manageTreeItemService.Move(ctx, request.LocalId, request.Id, request.NewParentId, request.Index)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, entry.ID(), "")
}

func (m *mcpTreeEdit) duplicate(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.MCPToolResult, error) {
	if entry.IsExample() {
		return m.duplicateExample(ctx, call, entry)
	}

	request := dto.NewTreeItemReference(call.LocalId, entry.ID())
	err := validateDuplicateItem(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	node, err := m.manageTreeItemService.Duplicate(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewTreeNode(*node).Id, "")
}

func (m *mcpTreeEdit) update(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	switch {
	case entry.IsExample():
		return m.updateExample(ctx, call, entry, arguments)
	case entry.IsRequest():
		return m.updateRequest(ctx, call, entry, arguments)
	default:
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditUpdateReason)
	}
}

func (m *mcpTreeEdit) edited(ctx context.Context, call dto.MCPToolCall, id, method string) (dto.MCPToolResult, error) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeEdited(entry, method)).WithTarget(entry.Target()), nil
}
