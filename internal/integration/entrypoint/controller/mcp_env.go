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
	mcpEnvListAction    = "list"
	mcpEnvGetAction     = "get"
	mcpEnvInspectAction = "inspect"
)

var envAITool = entity.AITool{
	Name:            "env",
	Description:     "List environments and read or inspect variables.",
	Actions:         []string{mcpEnvListAction, mcpEnvGetAction, mcpEnvInspectAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpEnv struct {
	errorHandler               adapter.ErrorHandler
	manageEnvironmentService   appAdapter.ManageEnvironmentService
	manageVariableTableService appAdapter.ManageVariableTableService
	searchTreeService          appAdapter.SearchTreeService
	saveRequestService         appAdapter.SaveRequestService
	inspectVariablesService    appAdapter.InspectVariablesService
}

func NewMCPEnv(
	errorHandler adapter.ErrorHandler,
	manageEnvironmentService appAdapter.ManageEnvironmentService,
	manageVariableTableService appAdapter.ManageVariableTableService,
	searchTreeService appAdapter.SearchTreeService,
	saveRequestService appAdapter.SaveRequestService,
	inspectVariablesService appAdapter.InspectVariablesService,
) adapter.MCPToolController {
	return &mcpEnv{
		errorHandler:               errorHandler,
		manageEnvironmentService:   manageEnvironmentService,
		manageVariableTableService: manageVariableTableService,
		searchTreeService:          searchTreeService,
		saveRequestService:         saveRequestService,
		inspectVariablesService:    inspectVariablesService,
	}
}

func (m *mcpEnv) Tool() dto.MCPTool {
	return dto.NewMCPTool(envAITool)
}

func (m *mcpEnv) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpEnvListAction:
		result, err = m.list(ctx, call)
	case mcpEnvGetAction:
		result, err = m.get(ctx, call)
	case mcpEnvInspectAction:
		result, err = m.inspect(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpEnv) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	request := dto.NewListEnvironments(call.LocalId)
	err := validateListEnvironments(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	list, err := m.manageEnvironmentService.List(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvList(*list)), nil
}

func (m *mcpEnv) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvScope
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.VariableTableReference(call.LocalId)
	err = validateGetVariableTable(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	table, err := m.manageVariableTableService.GetTable(ctx, request.LocalId, request.ScopeRef())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	name := m.scopeName(ctx, call.LocalId, table.Scope)
	return dto.NewMCPToolResult(dto.NewMCPEnvScopeVars(name, dto.NewVariableTable(*table))).WithTarget(table.Scope.Target(name)), nil
}

func (m *mcpEnv) scopeName(ctx context.Context, localId string, scope entity.ScopeRef) string {
	switch {
	case scope.IsEnvironment():
		if name, found := m.environmentName(ctx, localId, scope.ID); found {
			return name
		}
	case scope.IsCollection():
		if name, found := m.collectionName(ctx, localId, scope.ID); found {
			return name
		}
	}
	return scope.DisplayName()
}

func (m *mcpEnv) environmentName(ctx context.Context, localId, id string) (string, bool) {
	request := dto.NewEnvironmentReference(localId, id)
	err := validateGetEnvironment(&request)
	if err != nil {
		m.errorHandler.Handle(ctx, err, request)
		return "", false
	}

	details, err := m.manageEnvironmentService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		m.errorHandler.Handle(ctx, err, request)
		return "", false
	}
	return details.Environment.Name, true
}

func (m *mcpEnv) collectionName(ctx context.Context, localId, id string) (string, bool) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, localId, id)
	if err != nil {
		m.errorHandler.Handle(ctx, err, id)
		return "", false
	}
	return entry.Name(), true
}

func (m *mcpEnv) inspect(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvInspect
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	inspected, target, err := m.inspectedRequest(ctx, call, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := dto.NewInspectVariables(call.LocalId, inspected)
	err = validateInspectVariables(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	inspection, err := m.inspectVariablesService.Inspect(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvInspection(dto.NewVariableInspection(*inspection))).WithTarget(target), nil
}

func (m *mcpEnv) inspectedRequest(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPEnvInspect) (dto.Request, entity.AIActivityTarget, error) {
	if !arguments.HasRequest() {
		return arguments.Request(), entity.AIActivityTarget{}, nil
	}

	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.RequestItemId())
	if err != nil {
		return dto.Request{}, entity.AIActivityTarget{}, err
	}

	request := dto.NewOpenRequest(call.LocalId, entry.ID())
	err = validateOpenRequest(&request)
	if err != nil {
		return dto.Request{}, entity.AIActivityTarget{}, err
	}

	saved, err := m.saveRequestService.Open(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.Request{}, entity.AIActivityTarget{}, err
	}
	return dto.NewRequest(*saved), entry.Target(), nil
}
