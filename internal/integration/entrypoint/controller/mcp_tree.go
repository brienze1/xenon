package controller

import (
	"context"
	"strconv"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	mcpTreeListAction     = "list"
	mcpTreeGetAction      = "get"
	mcpTreeSettingsAction = "settings"
	mcpTreeSearchAction   = "search"
	mcpTreeTokenAction    = "token"

	mcpTreeDepthReason    = "depth must be an integer of 1 or more, or \"all\""
	mcpTreeIdReason       = "id is required"
	mcpTreeOwnerReason    = "owner_id is required"
	mcpTreeQueryReason    = "query is required"
	mcpTreeSettingsReason = "settings belong to a collection or a folder"
)

var treeAITool = entity.AITool{
	Name:            "tree",
	Description:     "Browse, read and search collections, folders, requests and examples.",
	Actions:         []string{mcpTreeListAction, mcpTreeGetAction, mcpTreeSettingsAction, mcpTreeSearchAction, mcpTreeTokenAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpTree struct {
	errorHandler               adapter.ErrorHandler
	searchTreeService          appAdapter.SearchTreeService
	saveRequestService         appAdapter.SaveRequestService
	manageItemSettingsService  appAdapter.ManageItemSettingsService
	manageExampleService       appAdapter.ManageExampleService
	manageVariableTableService appAdapter.ManageVariableTableService
	manageOAuthService         appAdapter.ManageOAuthService
}

func NewMCPTree(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	saveRequestService appAdapter.SaveRequestService,
	manageItemSettingsService appAdapter.ManageItemSettingsService,
	manageExampleService appAdapter.ManageExampleService,
	manageVariableTableService appAdapter.ManageVariableTableService,
	manageOAuthService appAdapter.ManageOAuthService,
) adapter.MCPToolController {
	return &mcpTree{
		errorHandler:               errorHandler,
		searchTreeService:          searchTreeService,
		saveRequestService:         saveRequestService,
		manageItemSettingsService:  manageItemSettingsService,
		manageExampleService:       manageExampleService,
		manageVariableTableService: manageVariableTableService,
		manageOAuthService:         manageOAuthService,
	}
}

func (m *mcpTree) Tool() dto.MCPTool {
	return dto.NewMCPTool(treeAITool)
}

func (m *mcpTree) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpTreeListAction:
		result, err = m.list(ctx, call)
	case mcpTreeGetAction:
		result, err = m.get(ctx, call)
	case mcpTreeSettingsAction:
		result, err = m.settings(ctx, call)
	case mcpTreeSearchAction:
		result, err = m.search(ctx, call)
	case mcpTreeTokenAction:
		result, err = m.token(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpTree) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPTreeList
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	depth, valid := arguments.ListDepth()
	if !valid {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeDepthReason)
	}

	tree, err := loadMCPTree(ctx, m.searchTreeService, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.HasParent() {
		return m.page(call, tree.Collections(), nil, depth, arguments.PageCursor(), entity.AIActivityTarget{})
	}

	parent, found := tree.Entry(arguments.ParentItemId())
	if !found || parent.IsExample() {
		return dto.MCPToolResult{}, mcpErrors.RejectMissingAIItem(arguments.ParentItemId())
	}
	return m.page(call, parent.Node.Children, parent.Node.Examples, depth, arguments.PageCursor(), parent.Target())
}

func (m *mcpTree) page(call dto.MCPToolCall, nodes []entity.TreeNode, examples []entity.TreeExample, depth int, cursor string, target entity.AIActivityTarget) (dto.MCPToolResult, error) {
	page, valid := dto.NewMCPTreeLinePage(nodes, examples, depth, cursor)
	if !valid {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpUnknownCursorReason)
	}
	return dto.NewMCPToolResult(page).WithTarget(target), nil
}

func (m *mcpTree) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	entry, err := m.item(ctx, call)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	switch {
	case entry.IsExample():
		return m.example(ctx, call, entry)
	case entry.IsRequest():
		return m.request(ctx, call, entry)
	default:
		return m.container(ctx, call, entry)
	}
}

func (m *mcpTree) request(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.MCPToolResult, error) {
	request := dto.NewOpenRequest(call.LocalId, entry.ID())
	err := validateOpenRequest(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	saved, err := m.saveRequestService.Open(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeRequest(dto.NewRequest(*saved), entry.Path)).WithTarget(entry.Target()), nil
}

func (m *mcpTree) container(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.MCPToolResult, error) {
	settings, err := m.itemSettings(ctx, call, entry)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeContainer(entry, settings)).WithTarget(entry.Target()), nil
}

func (m *mcpTree) example(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.MCPToolResult, error) {
	request := dto.NewExampleReference(call.LocalId, entry.RequestID, entry.ID())
	err := validateGetExample(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	details, err := m.manageExampleService.Get(ctx, request.LocalId, request.RequestId, request.ExampleId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeExample(dto.NewExample(*details))).WithTarget(entry.Target()), nil
}

func (m *mcpTree) settings(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	entry, err := m.item(ctx, call)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !entry.IsContainer() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeSettingsReason)
	}

	itemSettings, err := m.itemSettings(ctx, call, entry)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	settings := dto.NewMCPTreeSettings(itemSettings)
	if !entry.IsCollection() {
		return dto.NewMCPToolResult(settings).WithTarget(entry.Target()), nil
	}

	table, err := m.variableTable(ctx, call, entry)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(settings.WithVariables(table)).WithTarget(entry.Target()), nil
}

func (m *mcpTree) itemSettings(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.ItemSettings, error) {
	request := dto.NewTreeItemReference(call.LocalId, entry.ID())
	err := validateGetItemSettings(&request)
	if err != nil {
		return dto.ItemSettings{}, err
	}

	settings, err := m.manageItemSettingsService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.ItemSettings{}, err
	}
	return dto.NewItemSettings(*settings), nil
}

func (m *mcpTree) variableTable(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.VariableTable, error) {
	request := dto.NewVariableTableReference(call.LocalId, entity.NewCollectionScopeRef(entry.ID()))
	err := validateGetVariableTable(&request)
	if err != nil {
		return dto.VariableTable{}, err
	}

	table, err := m.manageVariableTableService.GetTable(ctx, request.LocalId, request.ScopeRef())
	if err != nil {
		return dto.VariableTable{}, err
	}
	return dto.NewVariableTable(*table), nil
}

func (m *mcpTree) search(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPTreeSearch
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.HasQuery() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeQueryReason)
	}
	if !arguments.HasValidLimit() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, "limit must be between 1 and "+strconv.Itoa(entity.AIPageSize))
	}

	request := arguments.SearchTree(call.LocalId)
	err = validateSearchTree(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	search, err := m.searchTreeService.Search(ctx, request.LocalId, request.Query)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeSearchList(search.Tree.Search(arguments.SearchQuery(), search.MatchIDs, arguments.ResultLimit()))), nil
}

func (m *mcpTree) token(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPTreeTokenOwner
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.HasOwner() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeOwnerReason)
	}

	owner, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.OwnerItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.OAuthTokenReference(call.LocalId)
	err = validateGetOAuthToken(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	token, err := m.manageOAuthService.Get(ctx, request.LocalId, request.OwnerId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if token == nil {
		return dto.MCPToolResult{}, usecase.RejectMissingOAuthToken()
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeToken(dto.NewTokenInfo(*token, timeutils.Now()))).WithTarget(owner.Target()), nil
}

func (m *mcpTree) item(ctx context.Context, call dto.MCPToolCall) (entity.TreeEntry, error) {
	var arguments dto.MCPTreeItem
	err := call.Decode(&arguments)
	if err != nil {
		return entity.TreeEntry{}, err
	}
	if !arguments.HasId() {
		return entity.TreeEntry{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeIdReason)
	}
	return findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.ItemId())
}
