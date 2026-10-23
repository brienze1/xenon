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
	mcpRunsGetAction  = "get"
	mcpRunsListAction = "list"
)

var runsAITool = entity.AITool{
	Name:            "runs",
	Description:     "Read a run or list the stored runs of a collection or folder.",
	Actions:         []string{mcpRunsGetAction, mcpRunsListAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpRuns struct {
	errorHandler    adapter.ErrorHandler
	findRunsService appAdapter.FindRunsService
}

func NewMCPRuns(errorHandler adapter.ErrorHandler, findRunsService appAdapter.FindRunsService) adapter.MCPToolController {
	return &mcpRuns{errorHandler: errorHandler, findRunsService: findRunsService}
}

func (m *mcpRuns) Tool() dto.MCPTool {
	return dto.NewMCPTool(
		runsAITool,
		dto.NewMCPToolArguments[dto.MCPRunReference](mcpRunsGetAction),
		dto.NewMCPToolArguments[dto.MCPRunTarget](mcpRunsListAction),
	)
}

func (m *mcpRuns) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpRunsGetAction:
		result, err = m.get(ctx, call)
	case mcpRunsListAction:
		result, err = m.list(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result.AsRecorded(), err
	}

	logger.Info(ctx, "Finished", result)
	return result.AsRecorded(), nil
}

func (m *mcpRuns) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
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

	run, err := m.findRunsService.Get(ctx, request.LocalId, request.RunId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewRun(*run)).WithTarget(entity.NewAIItemTarget(run.Target.ItemID)), nil
}

func (m *mcpRuns) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPRunTarget
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	err = validateMCPRunTarget(call, arguments)
	if err != nil {
		return dto.NewMCPTargetResult(entity.AIActivityTarget{}), err
	}

	runs, target, err := m.findRunsService.ListForAI(ctx, call.LocalId, arguments.Path(), arguments.TargetItemId())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	return dto.NewMCPToolResult(dto.NewRunSummaryViews(runs)).WithTarget(target), nil
}
