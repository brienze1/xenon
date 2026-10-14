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
	mcpExportOptionsAction      = "options"
	mcpExportRunAction          = "run"
	mcpExportResponseBodyAction = "response_body"
	mcpExportSecretsAction      = "Exporting secrets"
)

var exportAITool = entity.AITool{
	Name:            "export",
	Description:     "Export an item, environment or the workspace to a file path (never with secrets), or save a response body.",
	Actions:         []string{mcpExportOptionsAction, mcpExportRunAction, mcpExportResponseBodyAction},
	Effect:          aitooleffect.Destructive,
	WorkspaceScoped: true,
}

type mcpExport struct {
	errorHandler             adapter.ErrorHandler
	exportService            appAdapter.ExportService
	manageResponseService    appAdapter.ManageResponseService
	guardAIOutputService     appAdapter.GuardAIOutputService
	writeAIOutputFileService appAdapter.WriteAIOutputFileService
}

func NewMCPExport(
	errorHandler adapter.ErrorHandler,
	exportService appAdapter.ExportService,
	manageResponseService appAdapter.ManageResponseService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	writeAIOutputFileService appAdapter.WriteAIOutputFileService,
) adapter.MCPToolController {
	return &mcpExport{
		errorHandler:             errorHandler,
		exportService:            exportService,
		manageResponseService:    manageResponseService,
		guardAIOutputService:     guardAIOutputService,
		writeAIOutputFileService: writeAIOutputFileService,
	}
}

func (m *mcpExport) Tool() dto.MCPTool {
	return dto.NewMCPTool(exportAITool)
}

func (m *mcpExport) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpExportOptionsAction:
		result, err = m.options(ctx, call)
	case mcpExportRunAction:
		result, err = m.run(ctx, call)
	case mcpExportResponseBodyAction:
		result, err = m.responseBody(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpExport) options(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPExport
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	options, err := m.findOptions(ctx, call, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewExportOptions(*options)).WithTarget(arguments.ActivityTarget()), nil
}

func (m *mcpExport) run(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPExport
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if arguments.RequestsSecrets() {
		return dto.MCPToolResult{}, mcpErrors.RejectHumanOnlyAIAction(mcpExportSecretsAction)
	}

	options, err := m.findOptions(ctx, call, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.Export(call.LocalId, *options)
	err = validateExport(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	summary, err := m.exportService.Export(ctx, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewExportSummary(*summary)).WithTarget(arguments.ActivityTarget()), nil
}

func (m *mcpExport) findOptions(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPExport) (*entity.ExportOptions, error) {
	request := arguments.ExportItemReference(call.LocalId)
	err := validateExportOptions(&request)
	if err != nil {
		return nil, err
	}
	return m.exportService.Options(ctx, request.LocalId, request.Kind(), request.ItemId)
}
