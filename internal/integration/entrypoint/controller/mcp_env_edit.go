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
	mcpEnvEditCreateAction    = "create"
	mcpEnvEditUpdateAction    = "update"
	mcpEnvEditDuplicateAction = "duplicate"
	mcpEnvEditReorderAction   = "reorder"
	mcpEnvEditSetActiveAction = "set_active"
	mcpEnvEditSetVarsAction   = "set_vars"
	mcpEnvEditResetAction     = "reset"
	mcpEnvEditPersistAction   = "persist"
)

var envEditAITool = entity.AITool{
	Name:        entity.AIEnvEditToolName,
	Description: "Create, update or reorder environments, switch the active one, and edit variables (secret values are write-only).",
	Actions: []string{
		mcpEnvEditCreateAction, mcpEnvEditUpdateAction, mcpEnvEditDuplicateAction, mcpEnvEditReorderAction, mcpEnvEditSetActiveAction,
		entity.AIEnvEditAddVarAction, mcpEnvEditSetVarsAction, entity.AIEnvEditSetCurrentAction, mcpEnvEditResetAction, mcpEnvEditPersistAction,
	},
	Effect:          aitooleffect.Destructive,
	WorkspaceScoped: true,
}

type mcpEnvEdit struct {
	errorHandler                adapter.ErrorHandler
	manageEnvironmentService    appAdapter.ManageEnvironmentService
	autoSyncService             appAdapter.AutoSyncService
	setActiveEnvironmentService appAdapter.SetActiveEnvironmentService
	manageVariableTableService  appAdapter.ManageVariableTableService
}

func NewMCPEnvEdit(
	errorHandler adapter.ErrorHandler,
	manageEnvironmentService appAdapter.ManageEnvironmentService,
	autoSyncService appAdapter.AutoSyncService,
	setActiveEnvironmentService appAdapter.SetActiveEnvironmentService,
	manageVariableTableService appAdapter.ManageVariableTableService,
) adapter.MCPToolController {
	return &mcpEnvEdit{
		errorHandler:                errorHandler,
		manageEnvironmentService:    manageEnvironmentService,
		autoSyncService:             autoSyncService,
		setActiveEnvironmentService: setActiveEnvironmentService,
		manageVariableTableService:  manageVariableTableService,
	}
}

func (m *mcpEnvEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(envEditAITool)
}

func (m *mcpEnvEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpEnvEditCreateAction:
		result, err = m.create(ctx, call)
	case mcpEnvEditUpdateAction:
		result, err = m.update(ctx, call)
	case mcpEnvEditDuplicateAction:
		result, err = m.duplicate(ctx, call)
	case mcpEnvEditReorderAction:
		result, err = m.reorder(ctx, call)
	case mcpEnvEditSetActiveAction:
		result, err = m.setActive(ctx, call)
	case entity.AIEnvEditAddVarAction:
		result, err = m.addVariable(ctx, call)
	case mcpEnvEditSetVarsAction:
		result, err = m.setVariables(ctx, call)
	case entity.AIEnvEditSetCurrentAction:
		result, err = m.setCurrent(ctx, call)
	case mcpEnvEditResetAction:
		result, err = m.reset(ctx, call)
	case mcpEnvEditPersistAction:
		result, err = m.persist(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	result = result.WithChanges(aichangekind.Variables)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpEnvEdit) create(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvEdit
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.CreateEnvironment(call.LocalId)
	err = validateCreateEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	created, err := m.manageEnvironmentService.Create(ctx, request.LocalId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.EditsSettings() {
		return editedEnvironment(created.Environment), nil
	}
	return m.save(ctx, call, created.Environment, arguments.WithoutName())
}

func (m *mcpEnvEdit) update(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvEdit
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.EnvironmentReference(call.LocalId)
	err = validateGetEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	stored, err := m.manageEnvironmentService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.save(ctx, call, stored.Environment, arguments)
}

func (m *mcpEnvEdit) save(ctx context.Context, call dto.MCPToolCall, stored entity.Environment, arguments dto.MCPEnvEdit) (dto.MCPToolResult, error) {
	request := arguments.UpdateEnvironment(call.LocalId, stored)
	err := validateUpdateEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	updated, err := m.manageEnvironmentService.Update(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Saved, request.LocalId)
	return editedEnvironment(updated.Environment), nil
}

func (m *mcpEnvEdit) duplicate(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvEdit
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.EnvironmentReference(call.LocalId)
	err = validateDuplicateEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	duplicated, err := m.manageEnvironmentService.Duplicate(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return editedEnvironment(duplicated.Environment), nil
}

func (m *mcpEnvEdit) reorder(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvEdit
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.ReorderEnvironment(call.LocalId)
	err = validateReorderEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	list, err := m.manageEnvironmentService.Reorder(ctx, request.LocalId, request.Id, request.Index)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvList(*list)), nil
}

func (m *mcpEnvEdit) setActive(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvActivation
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.SetActiveEnvironment(call.LocalId)
	err = validateSetActiveEnvironment(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	activeId, err := m.setActiveEnvironmentService.SetActive(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvActive(activeId)).WithTarget(entity.NewAIEnvironmentTarget(activeId)), nil
}

func (m *mcpEnvEdit) addVariable(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvVariable
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.AddVariable(call.LocalId)
	err = validateAddVariable(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entry, err := m.manageVariableTableService.AddVariable(ctx, request.LocalId, request.ScopeRef(), request.ToEntity(), request.Value)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPVariableLine(dto.NewVariableRow(*entry))).WithTarget(arguments.Target()), nil
}

func (m *mcpEnvEdit) setVariables(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvVariables
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	table, err := saveMCPVariableRows(ctx, m.manageVariableTableService, call, arguments.VariableScope(), arguments.VariableRows())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvScopeVariables(table)).WithTarget(arguments.Target()), nil
}

func (m *mcpEnvEdit) setCurrent(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvVariable
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.SetCurrentValue(call.LocalId)
	err = validateSetCurrentValue(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entry, err := m.manageVariableTableService.SetCurrent(ctx, request.LocalId, request.ScopeRef(), request.Key, request.Value)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPVariableLine(dto.NewVariableRow(*entry))).WithTarget(arguments.Target()), nil
}

func (m *mcpEnvEdit) reset(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvVariable
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.VariableKeys(call.LocalId)
	err = validateResetVariables(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	table, err := m.manageVariableTableService.Reset(ctx, request.LocalId, request.ScopeRef(), request.Keys)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvScopeVariables(dto.NewVariableTable(*table))).WithTarget(arguments.Target()), nil
}

func (m *mcpEnvEdit) persist(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPEnvVariable
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.VariableKeys(call.LocalId)
	err = validatePersistVariables(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	table, err := m.manageVariableTableService.Persist(ctx, request.LocalId, request.ScopeRef(), request.Keys)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEnvScopeVariables(dto.NewVariableTable(*table))).WithTarget(arguments.Target()), nil
}

func editedEnvironment(environment entity.Environment) dto.MCPToolResult {
	return dto.NewMCPToolResult(dto.NewMCPEnvLine(environment, false)).WithTarget(environment.Target())
}
