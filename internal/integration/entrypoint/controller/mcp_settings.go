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

const mcpSettingsGetAction = "get"

var settingsAITool = entity.AITool{
	Name:            "settings",
	Description:     "Read app settings.",
	Actions:         []string{mcpSettingsGetAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpSettings struct {
	errorHandler             adapter.ErrorHandler
	manageAppSettingsService appAdapter.ManageAppSettingsService
}

func NewMCPSettings(errorHandler adapter.ErrorHandler, manageAppSettingsService appAdapter.ManageAppSettingsService) adapter.MCPToolController {
	return &mcpSettings{errorHandler: errorHandler, manageAppSettingsService: manageAppSettingsService}
}

func (m *mcpSettings) Tool() dto.MCPTool {
	return dto.NewMCPTool(settingsAITool)
}

func (m *mcpSettings) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpSettingsGetAction:
		result, err = m.get(ctx)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpSettings) get(ctx context.Context) (dto.MCPToolResult, error) {
	settings, err := m.manageAppSettingsService.Get(ctx)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPAppSettings(*settings)), nil
}
