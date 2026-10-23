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
	errorHandler      adapter.ErrorHandler
	startRunService   appAdapter.StartRunService
	controlRunService appAdapter.ControlRunService
}

func NewMCPRun(errorHandler adapter.ErrorHandler, startRunService appAdapter.StartRunService, controlRunService appAdapter.ControlRunService) adapter.MCPToolController {
	return &mcpRun{errorHandler: errorHandler, startRunService: startRunService, controlRunService: controlRunService}
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

	err = validateMCPRunTarget(call, arguments.MCPRunTarget)
	if err != nil {
		return dto.NewMCPTargetResult(entity.AIActivityTarget{}), err
	}
	err = validateMCPRun(&arguments)
	if err != nil {
		return dto.NewMCPTargetResult(entity.AIActivityTarget{}), err
	}

	started, err := m.startRunService.StartForAI(ctx, call.LocalId, arguments.Path(), arguments.TargetItemId(), arguments.ToEntity(), properties.Properties().Application.Version, call.Client, call.Session)
	if err != nil {
		return dto.NewMCPTargetResult(started.Target), err
	}
	if started.HasUnknownEnvironment() {
		return dto.NewMCPTargetResult(started.Target), mcpErrors.RejectInvalidArguments(call.Tool, mcpRunEnvironmentNotFound)
	}
	if started.IsRefused() {
		return dto.NewMCPTargetResult(started.Target), rejectRefusedAIApproval(*started.Refused)
	}
	return dto.NewMCPToolResult(dto.NewMCPRunStarted(*started.Run)).WithTarget(started.Target).WithChanges(started.Changes...), nil
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

	run, err := m.controlRunService.Stop(ctx, request.LocalId, request.RunId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewRunSummaryView(*run)).WithTarget(entity.NewAIItemTarget(run.Target.ItemID)), nil
}
