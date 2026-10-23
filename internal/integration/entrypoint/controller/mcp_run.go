package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpRunStartAction         = "start"
	mcpRunStopAction          = "stop"
	mcpRunTargetConflict      = "pass target or item_id, not both"
	mcpRunTargetMissing       = "target or item_id is required"
	mcpRunIdMissing           = "run_id is required"
	mcpRunEnvironmentNotFound = "environment must be an environment name or id, or none"
)

var runAITool = entity.AITool{
	Name:            "run",
	Description:     "Start or stop a collection or folder run.",
	Actions:         []string{mcpRunStartAction, mcpRunStopAction},
	Effect:          aitooleffect.Destructive,
	OpenWorld:       true,
	WorkspaceScoped: true,
}

type mcpRun struct {
	errorHandler             adapter.ErrorHandler
	searchTreeService        appAdapter.SearchTreeService
	manageRunnerService      appAdapter.ManageRunnerService
	manageAppSettingsService appAdapter.ManageAppSettingsService
	approveAISendService     appAdapter.ApproveAISendService
}

func NewMCPRun(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	manageRunnerService appAdapter.ManageRunnerService,
	manageAppSettingsService appAdapter.ManageAppSettingsService,
	approveAISendService appAdapter.ApproveAISendService,
) adapter.MCPToolController {
	return &mcpRun{
		errorHandler:             errorHandler,
		searchTreeService:        searchTreeService,
		manageRunnerService:      manageRunnerService,
		manageAppSettingsService: manageAppSettingsService,
		approveAISendService:     approveAISendService,
	}
}

func (m *mcpRun) Tool() dto.MCPTool {
	return dto.NewMCPTool(
		runAITool,
		dto.NewMCPToolArguments[dto.MCPRun](mcpRunStartAction),
		dto.NewMCPToolArguments[dto.MCPRunReference](mcpRunStopAction),
	)
}

func (m *mcpRun) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpRunStartAction:
		result, err = m.start(ctx, call)
	case mcpRunStopAction:
		result, err = m.stop(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result.AsRecorded(), err
	}

	logger.Info(ctx, "Finished", result)
	return result.AsRecorded(), nil
}

func (m *mcpRun) start(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPRun
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	itemId, target, err := resolveMCPRunTarget(ctx, m.searchTreeService, call, arguments.RunTarget())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	plan, err := m.plan(ctx, call.LocalId, itemId)
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	config, known := arguments.RunConfig(*plan)
	if !known {
		return dto.NewMCPTargetResult(target), mcpErrors.RejectInvalidArguments(call.Tool, mcpRunEnvironmentNotFound)
	}

	run, err := m.startRun(ctx, arguments.StartRun(call.LocalId, itemId, "", config))
	if environment, unconfirmed := mcpErrors.UnconfirmedSendEnvironment(err); unconfirmed {
		run, err = m.startConfirmed(ctx, call, arguments.StartRun(call.LocalId, itemId, environment.ID, config), plan.Target.Name, environment)
	}
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	return dto.NewMCPToolResult(dto.NewMCPRunStarted(*run)).WithTarget(target).WithChanges(arguments.Changes(config)...), nil
}

func (m *mcpRun) stop(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPRunReference
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.HasRunId() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpRunIdMissing)
	}

	request := arguments.RunReference(call.LocalId)
	err = validateRunReference(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	run, err := m.manageRunnerService.Stop(ctx, request.LocalId, request.RunId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewRunSummaryView(*run)).WithTarget(entity.NewAIItemTarget(run.Target.ItemID)), nil
}

func (m *mcpRun) plan(ctx context.Context, localId, itemId string) (*dto.RunPlan, error) {
	request := dto.RunTargetReference{LocalId: localId, ItemId: itemId}
	err := validateRunTargetReference(&request)
	if err != nil {
		return nil, err
	}

	plan, err := m.manageRunnerService.Plan(ctx, request.LocalId, request.ItemId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunPlan(*plan)
	return &result, nil
}

func (m *mcpRun) startConfirmed(ctx context.Context, call dto.MCPToolCall, request dto.StartRun, targetName string, environment entity.AIApprovalEnvironment) (*entity.Run, error) {
	settings, err := m.manageAppSettingsService.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.SkipsAIApprovals() {
		summary := entity.NewRunApprovalSummary(targetName, len(request.Config.ToEntity().SelectedRows()))
		err = askMCPApproval(ctx, m.approveAISendService, dto.NewAIRunApprovalRequest(call.Session, call.Client, summary, dto.NewAIApprovalEnvironment(environment)))
		if err != nil {
			return nil, err
		}
	}
	return m.startRun(ctx, request)
}

func (m *mcpRun) startRun(ctx context.Context, request dto.StartRun) (*entity.Run, error) {
	err := validateStartRun(&request)
	if err != nil {
		return nil, err
	}
	return m.manageRunnerService.Start(ctx, request.ToEntity(), properties.Properties().Application.Version)
}
