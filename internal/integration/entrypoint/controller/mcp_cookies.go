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
	mcpCookiesListAction      = "list"
	mcpCookiesAllowlistAction = "allowlist"
)

var cookiesAITool = entity.AITool{
	Name:            "cookies",
	Description:     "List cookies and the cookie allowlist.",
	Actions:         []string{mcpCookiesListAction, mcpCookiesAllowlistAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpCookies struct {
	errorHandler         adapter.ErrorHandler
	manageCookiesService appAdapter.ManageCookiesService
}

func NewMCPCookies(errorHandler adapter.ErrorHandler, manageCookiesService appAdapter.ManageCookiesService) adapter.MCPToolController {
	return &mcpCookies{errorHandler: errorHandler, manageCookiesService: manageCookiesService}
}

func (m *mcpCookies) Tool() dto.MCPTool {
	return dto.NewMCPTool(cookiesAITool)
}

func (m *mcpCookies) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpCookiesListAction:
		result, err = m.list(ctx, call)
	case mcpCookiesAllowlistAction:
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

func (m *mcpCookies) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPCookiesList
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := dto.NewCookieJarReference(call.LocalId)
	err = validateListCookies(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	domains, err := m.manageCookiesService.List(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	page, valid := arguments.Page(domains)
	if !valid {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpUnknownCursorReason)
	}
	return dto.NewMCPToolResult(dto.NewMCPPage(page, dto.NewMCPCookieLine)), nil
}

func (m *mcpCookies) allowlist(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	request := dto.NewCookieJarReference(call.LocalId)
	err := validateGetCookieAllowlist(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	allowlist, err := m.manageCookiesService.GetAllowlist(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewCookieAllowlist(*allowlist)), nil
}
