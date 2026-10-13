package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	deniedCode      = "MCP-03403"
	timedOutCode    = "MCP-04408"
	unknownCode     = "UNK-02500"
	unknownMessage  = "Unknown error"
	guardFailedText = `{"code":"MCP-12500","message":"The result could not be checked for secrets"}`
)

type toolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type pipeline struct {
	tools *tools
	tool  tool
}

func (p pipeline) handle(ctx context.Context, request *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
	started := time.Now()
	call := p.newCall(request)
	result, err := p.run(ctx, &call)
	activity := dto.AIActivityRecord{Tool: call.Tool, Action: call.Action, Client: call.Client, Args: string(call.arguments()), Target: result.Target, Outcome: entity.AIActivityOk}
	text, failure := p.respond(ctx, &call, result, err)
	if failure != nil {
		activity.Outcome, activity.Code = outcomeOf(failure.Code), failure.Code
	}
	activity.LocalId = p.recordedWorkspace(ctx, call)
	activity.DurationMs = time.Since(started).Milliseconds()
	logger.Info(ctx, "MCP call", map[string]any{"client": call.Client, "tool": call.Tool, "action": call.Action, "outcome": activity.Outcome, "code": activity.Code, "duration_ms": activity.DurationMs})
	if !result.Recorded {
		_ = p.tools.controllers.AIAccess.Record(ctx, activity)
	}
	if failure == nil && len(result.Changes) > 0 && call.LocalId != "" {
		_ = p.tools.controllers.AIAccess.Notify(ctx, dto.AIChanges{LocalId: call.LocalId, Changes: result.Changes})
	}
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: text}}, IsError: failure != nil}, nil
}

func (p pipeline) newCall(request *sdk.CallToolRequest) Call {
	call := Call{Tool: p.tool.name, Client: entity.AIClientName(clientName(request))}
	if request.Params != nil {
		call.Arguments = request.Params.Arguments
	}
	if request.Session != nil {
		call.Session = request.Session.ID()
	}
	return call
}

func (p pipeline) run(ctx context.Context, call *Call) (Result, error) {
	localId, err := p.activeWorkspace(ctx)
	if err != nil {
		return Result{}, err
	}
	call.LocalId = localId
	action, err := p.action(*call)
	call.Action = action
	if err != nil {
		return Result{}, err
	}
	if p.tool.workspaceScoped && call.LocalId == "" {
		return Result{}, noActiveWorkspace()
	}
	call.guard = func(ctx context.Context, text string) string {
		return p.guard(ctx, localId, text)
	}
	return p.tool.handler(p.tools)(ctx, *call)
}

func (p pipeline) action(call Call) (string, error) {
	if !p.tool.hasActions() {
		return "", nil
	}
	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(call.arguments(), &arguments); err != nil {
		return "", invalidArguments(call.Tool, "the arguments must be a JSON object")
	}
	var action string
	if raw, found := arguments[p.tool.argumentName()]; found {
		if err := json.Unmarshal(raw, &action); err != nil {
			return string(raw), unknownAction(call.Tool, string(raw), p.tool.actions)
		}
	}
	if action == "" {
		action = p.tool.defaultAction()
	}
	if !p.tool.allows(action) {
		return action, unknownAction(call.Tool, action, p.tool.actions)
	}
	return action, nil
}

func (p pipeline) respond(ctx context.Context, call *Call, result Result, err error) (string, *toolError) {
	if err != nil {
		failure := toolErrorOf(err)
		encoded, _ := encodeJSON(failure)
		return p.guard(ctx, call.LocalId, encoded), &failure
	}
	encoded, err := encodeJSON(result.Payload)
	if err != nil {
		failure := toolErrorOf(err)
		encoded, _ = encodeJSON(failure)
		return p.guard(ctx, call.LocalId, encoded), &failure
	}
	return p.guard(ctx, call.LocalId, encoded), nil
}

func (p pipeline) guard(ctx context.Context, localId, text string) string {
	guarded, err := p.tools.controllers.AIAccess.Guard(ctx, dto.AIGuard{LocalId: localId, Text: text})
	if err != nil {
		return guardFailedText
	}
	return guarded.Text
}

func (p pipeline) activeWorkspace(ctx context.Context) (string, error) {
	workspaces, err := p.tools.controllers.WorkspaceList.Handle(ctx)
	if err != nil {
		return "", err
	}
	return workspaces.ActiveLocalId, nil
}

func (p pipeline) recordedWorkspace(ctx context.Context, call Call) string {
	if p.tool.workspaceScoped {
		return call.LocalId
	}
	localId, err := p.activeWorkspace(ctx)
	if err != nil {
		return call.LocalId
	}
	return localId
}

func toolErrorOf(err error) toolError {
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		return toolError{Code: unknownCode, Message: unknownMessage}
	}
	return toolError{Code: custom.Code(), Message: custom.Description()}
}

func outcomeOf(code string) string {
	switch code {
	case deniedCode:
		return entity.AIActivityDenied
	case timedOutCode:
		return entity.AIActivityTimedOut
	default:
		return entity.AIActivityError
	}
}

func clientName(request *sdk.CallToolRequest) string {
	if info := request.ClientInfo(); info != nil {
		return info.Name
	}
	return ""
}
