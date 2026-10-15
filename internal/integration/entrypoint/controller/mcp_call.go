package controller

import (
	"context"
	"errors"
	"time"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const mcpNotObjectReason = "the arguments must be a JSON object"

type mcpCall struct {
	errorHandler            adapter.ErrorHandler
	listWorkspacesService   appAdapter.ListWorkspacesService
	guardAIOutputService    appAdapter.GuardAIOutputService
	recordAIActivityService appAdapter.RecordAIActivityService
	notifyAIChangeService   appAdapter.NotifyAIChangeService
	approveAISendService    appAdapter.ApproveAISendService
}

func NewMCPCall(
	errorHandler adapter.ErrorHandler,
	listWorkspacesService appAdapter.ListWorkspacesService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	recordAIActivityService appAdapter.RecordAIActivityService,
	notifyAIChangeService appAdapter.NotifyAIChangeService,
	approveAISendService appAdapter.ApproveAISendService,
) adapter.MCPCallController {
	return &mcpCall{
		errorHandler:            errorHandler,
		listWorkspacesService:   listWorkspacesService,
		guardAIOutputService:    guardAIOutputService,
		recordAIActivityService: recordAIActivityService,
		notifyAIChangeService:   notifyAIChangeService,
		approveAISendService:    approveAISendService,
	}
}

func (m *mcpCall) Call(mcpCtx context.Context, call dto.MCPCall, tool adapter.MCPToolController) (response dto.MCPOutcome) {
	ctx := logger.GetContext(mcpCtx)
	logger.Info(ctx, "Started", call)

	defer func(ctx context.Context, call any) {
		if recovered := recover(); recovered != nil {
			response = m.panicked(ctx, call, recovered)
		}
	}(ctx, call)

	started := time.Now()
	definition := tool.Tool()
	spec := definition.ToEntity()
	toolCall, result, err := m.run(ctx, call, spec, tool)
	text, failure := m.respond(ctx, definition, toolCall, result, err)
	code := codeOf(failure)
	outcome := entity.AIActivityOutcomeOf(code).String()
	localId := m.recordedWorkspace(ctx, spec, toolCall.LocalId)
	durationMs := time.Since(started).Milliseconds()
	logger.Info(ctx, "MCP call", map[string]any{"client": call.Client, "tool": spec.Name, "action": toolCall.Action, "outcome": outcome, "code": code, "duration_ms": durationMs})
	if result.NeedsRecording() {
		m.record(ctx, dto.NewAIActivityRecord(localId, call.Client, spec.Name, toolCall.Action, result.Target, outcome, code, string(call.Args()), durationMs))
	}
	if result.NotifiesChanges(failure != nil, toolCall.LocalId) {
		m.notify(ctx, dto.NewAIChanges(toolCall.LocalId, result.Changes))
	}
	response = dto.NewMCPOutcome(text, failure != nil)

	logger.Info(ctx, "Finished", response)
	return response
}

func (m *mcpCall) CloseSession(mcpCtx context.Context, session dto.MCPSession) (err error) {
	ctx := logger.GetContext(mcpCtx)
	logger.Info(ctx, "Started", session)

	defer func(ctx context.Context, session any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, session, recovered)
		}
		if err != nil {
			err = m.errorHandler.Handle(ctx, err, session)
		}
	}(ctx, session)

	err = validateAIArguments(&session)
	if err != nil {
		return err
	}

	err = m.approveAISendService.Deny(ctx, session.SessionId())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", session)
	return nil
}

func (m *mcpCall) run(ctx context.Context, call dto.MCPCall, spec entity.AITool, tool adapter.MCPToolController) (toolCall dto.MCPToolCall, result dto.MCPToolResult, err error) {
	defer func(ctx context.Context, call any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, call, recovered)
		}
		if err != nil {
			err = m.errorHandler.Handle(ctx, err, call)
		}
	}(ctx, call)

	err = validateAIArguments(&call)
	if err != nil {
		return dto.NewMCPToolCall(call, spec.Name, "", ""), result, err
	}

	localId, err := m.activeWorkspace(ctx)
	if err != nil {
		return dto.NewMCPToolCall(call, spec.Name, "", ""), result, err
	}
	action, err := m.action(call, spec)
	toolCall = dto.NewMCPToolCall(call, spec.Name, action, localId)
	if err != nil {
		return toolCall, result, err
	}
	if spec.MissesWorkspace(localId) {
		return toolCall, result, mcpErrors.RejectInactiveAIWorkspace()
	}

	result, err = tool.Handle(ctx, toolCall)
	return toolCall, result, err
}

func (m *mcpCall) action(call dto.MCPCall, spec entity.AITool) (string, error) {
	if !spec.HasActions() {
		return "", nil
	}
	action, raw, isObject, isString := call.Selector(spec.ArgumentName())
	if !isObject {
		return "", mcpErrors.RejectInvalidArguments(spec.Name, mcpNotObjectReason)
	}
	if !isString {
		return raw, mcpErrors.RejectUnknownAIAction(spec.Name, raw, spec.Actions)
	}
	action = spec.ActionOrDefault(action)
	if !spec.Allows(action) {
		return action, mcpErrors.RejectUnknownAIAction(spec.Name, action, spec.Actions)
	}
	return action, nil
}

func (m *mcpCall) respond(ctx context.Context, definition dto.MCPTool, toolCall dto.MCPToolCall, result dto.MCPToolResult, err error) (string, errs.BaseError) {
	if err == nil {
		text, encodeErr := dto.EncodeMCPText(result.Payload)
		if encodeErr == nil {
			return m.guard(ctx, toolCall.LocalId, text), nil
		}
		err = m.errorHandler.Handle(ctx, mcpErrors.FailUnknownAICall(encodeErr), result)
	}
	failure := failureOf(err)
	text, encodeErr := dto.EncodeMCPText(dto.NewMCPToolError(failure, definition, toolCall.Action))
	if encodeErr != nil {
		m.errorHandler.Handle(ctx, mcpErrors.FailUnknownAICall(encodeErr), result)
		return mcpErrors.UnguardedAIOutputText, failure
	}
	return m.guard(ctx, toolCall.LocalId, text), failure
}

func (m *mcpCall) panicked(ctx context.Context, call any, recovered any) dto.MCPOutcome {
	failure := failureOf(m.errorHandler.Handle(ctx, handlePanic(ctx, call, recovered), call))
	text, err := dto.EncodeMCPText(dto.NewMCPPanicToolError(failure))
	if err != nil {
		m.errorHandler.Handle(ctx, mcpErrors.FailUnknownAICall(err), call)
		return dto.NewMCPOutcome(mcpErrors.UnguardedAIOutputText, true)
	}
	return dto.NewMCPOutcome(m.guard(ctx, "", text), true)
}

func (m *mcpCall) guard(ctx context.Context, localId, text string) string {
	return guardMCPText(ctx, m.errorHandler, m.guardAIOutputService, localId, text)
}

func (m *mcpCall) record(ctx context.Context, record dto.AIActivityRecord) {
	defer func(ctx context.Context, record any) {
		if recovered := recover(); recovered != nil {
			m.errorHandler.Handle(ctx, handlePanic(ctx, record, recovered), record)
		}
	}(ctx, record)

	if err := validateAIActivityRecord(&record); err != nil {
		m.errorHandler.Handle(ctx, err, record)
		return
	}
	if err := m.recordAIActivityService.Record(ctx, record.WorkspaceId(), record.ClientName(), record.ToEntity()); err != nil {
		m.errorHandler.Handle(ctx, err, record)
	}
}

func (m *mcpCall) notify(ctx context.Context, changes dto.AIChanges) {
	defer func(ctx context.Context, changes any) {
		if recovered := recover(); recovered != nil {
			m.errorHandler.Handle(ctx, handlePanic(ctx, changes, recovered), changes)
		}
	}(ctx, changes)

	if err := validateAIChanges(&changes); err != nil {
		m.errorHandler.Handle(ctx, err, changes)
		return
	}
	if err := m.notifyAIChangeService.Notify(ctx, changes.WorkspaceId(), changes.ToEntity()); err != nil {
		m.errorHandler.Handle(ctx, err, changes)
	}
}

func (m *mcpCall) recordedWorkspace(ctx context.Context, spec entity.AITool, localId string) string {
	if spec.IsWorkspaceScoped() {
		return localId
	}
	active, err := m.activeWorkspace(ctx)
	if err != nil {
		m.errorHandler.Handle(ctx, err, nil)
		return localId
	}
	return active
}

func (m *mcpCall) activeWorkspace(ctx context.Context) (string, error) {
	registry, err := m.listWorkspacesService.List(ctx)
	if err != nil {
		return "", err
	}
	return registry.ActiveLocalID, nil
}

func failureOf(err error) errs.BaseError {
	var failure errs.BaseError
	if !errors.As(err, &failure) {
		errors.As(mcpErrors.FailUnknownAICall(err), &failure)
	}
	return failure
}

func codeOf(failure errs.BaseError) string {
	if failure == nil {
		return ""
	}
	return failure.Code()
}
