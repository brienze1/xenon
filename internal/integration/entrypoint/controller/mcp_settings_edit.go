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
	mcpSettingsEditSaveAction = "save"
	mcpSettingsEditAIAccess   = "Changing AI access"
)

var settingsEditAITool = entity.AITool{
	Name:            "settings_edit",
	Description:     "Change app settings (script timeout, history caps).",
	Actions:         []string{mcpSettingsEditSaveAction},
	Effect:          aitooleffect.Write,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpSettingsEdit struct {
	errorHandler             adapter.ErrorHandler
	manageAppSettingsService appAdapter.ManageAppSettingsService
}

func NewMCPSettingsEdit(errorHandler adapter.ErrorHandler, manageAppSettingsService appAdapter.ManageAppSettingsService) adapter.MCPToolController {
	return &mcpSettingsEdit{errorHandler: errorHandler, manageAppSettingsService: manageAppSettingsService}
}

func (m *mcpSettingsEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(settingsEditAITool, dto.NewMCPToolArguments[dto.MCPSettingsSave](mcpSettingsEditSaveAction))
}

func (m *mcpSettingsEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpSettingsEditSaveAction:
		result, err = m.save(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpSettingsEdit) save(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPSettingsSave
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if arguments.ChangesAIAccess() {
		return dto.MCPToolResult{}, mcpErrors.RejectHumanOnlyAIAction(mcpSettingsEditAIAccess)
	}

	request := arguments.AppSettings()
	err = validateSaveSettings(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	saved, err := m.manageAppSettingsService.Save(ctx, request.ToEntity())
	if saved == nil {
		return dto.MCPToolResult{}, err
	}
	if err != nil {
		m.errorHandler.Handle(ctx, err, request)
	}
	return dto.NewMCPToolResult(dto.NewMCPAppSettings(*saved)), nil
}
