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
	mcpCookiesEditUpsertAction    = "upsert"
	mcpCookiesEditAllowlistAction = "allowlist"
)

var cookiesEditAITool = entity.AITool{
	Name:            "cookies_edit",
	Description:     "Add or edit cookies and the cookie allowlist.",
	Actions:         []string{mcpCookiesEditUpsertAction, mcpCookiesEditAllowlistAction},
	Effect:          aitooleffect.Destructive,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpCookiesEdit struct {
	errorHandler         adapter.ErrorHandler
	manageCookiesService appAdapter.ManageCookiesService
}

func NewMCPCookiesEdit(errorHandler adapter.ErrorHandler, manageCookiesService appAdapter.ManageCookiesService) adapter.MCPToolController {
	return &mcpCookiesEdit{errorHandler: errorHandler, manageCookiesService: manageCookiesService}
}

func (m *mcpCookiesEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(cookiesEditAITool)
}

func (m *mcpCookiesEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpCookiesEditUpsertAction:
		result, err = m.upsert(ctx, call)
	case mcpCookiesEditAllowlistAction:
		result, err = m.allowlist(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpCookiesEdit) upsert(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPCookieUpsert
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.UpsertCookie(call.LocalId)
	err = validateUpsertCookie(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	stored, err := m.manageCookiesService.Upsert(ctx, request.LocalId, request.Domain, request.Raw, request.ReplacedCookie())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPCookieLine(*stored)).WithTarget(stored.ActivityTarget()).WithChanges(aichangekind.Cookies), nil
}

func (m *mcpCookiesEdit) allowlist(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPCookieAllowlist
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.SaveCookieAllowlist(call.LocalId)
	err = validateSaveCookieAllowlist(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	allowlist, err := m.manageCookiesService.SaveAllowlist(ctx, request.LocalId, request.Domains)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewCookieAllowlist(*allowlist)).WithChanges(aichangekind.Cookies), nil
}
