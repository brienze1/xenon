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
	mcpImportDetectAction  = "detect"
	mcpImportClashesAction = "clashes"
	mcpImportRunAction     = "run"
	mcpImportDiscardAction = "discard"
)

var importAITool = entity.AITool{
	Name:            "import",
	Description:     "Import Postman v2 or v3 files and folders, Bruno or Xenon files from paths into the workspace.",
	Actions:         []string{mcpImportDetectAction, mcpImportClashesAction, mcpImportRunAction, mcpImportDiscardAction},
	Effect:          aitooleffect.Destructive,
	WorkspaceScoped: true,
}

type mcpImport struct {
	errorHandler        adapter.ErrorHandler
	manageImportService appAdapter.ManageImportService
}

func NewMCPImport(errorHandler adapter.ErrorHandler, manageImportService appAdapter.ManageImportService) adapter.MCPToolController {
	return &mcpImport{errorHandler: errorHandler, manageImportService: manageImportService}
}

func (m *mcpImport) Tool() dto.MCPTool {
	return dto.NewMCPTool(importAITool)
}

func (m *mcpImport) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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

	var arguments dto.MCPImport
	err = call.Decode(&arguments)
	if err != nil {
		return result, err
	}

	switch call.Action {
	case mcpImportDetectAction:
		result, err = m.detect(ctx, arguments)
	case mcpImportClashesAction:
		result, err = m.clashes(ctx, call, arguments)
	case mcpImportRunAction:
		result, err = m.run(ctx, call, arguments.WithDefaultClash())
	case mcpImportDiscardAction:
		result, err = m.discard(ctx, arguments)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpImport) detect(ctx context.Context, arguments dto.MCPImport) (dto.MCPToolResult, error) {
	source, err := m.detectSource(ctx, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewImportPreview(*source)), nil
}

func (m *mcpImport) clashes(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPImport) (dto.MCPToolResult, error) {
	source, err := m.detectSource(ctx, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	defer m.discardSource(ctx, source.ID)

	request := arguments.ImportPlan(call.LocalId, *source)
	err = validateImportClashes(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	clashes, err := m.manageImportService.Clashes(ctx, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPList(dto.NewImportClashes(clashes).Clashes)), nil
}

func (m *mcpImport) run(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPImport) (dto.MCPToolResult, error) {
	source, err := m.detectSource(ctx, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	defer m.discardSource(ctx, source.ID)

	request := arguments.ImportPlan(call.LocalId, *source)
	err = validateImportPlan(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	imported, err := m.manageImportService.Import(ctx, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewImportResult(*imported)).WithChanges(aichangekind.Tree, aichangekind.Variables), nil
}

func (m *mcpImport) discard(ctx context.Context, arguments dto.MCPImport) (dto.MCPToolResult, error) {
	if !arguments.HasSource() {
		return dto.NewMCPToolResult(dto.NewMCPEmpty()), nil
	}

	err := m.discardImport(ctx, arguments.Discard())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPEmpty()), nil
}

func (m *mcpImport) detectSource(ctx context.Context, arguments dto.MCPImport) (*entity.ImportSource, error) {
	request := arguments.DetectImport()
	err := validateDetectImport(&request)
	if err != nil {
		return nil, err
	}
	return m.manageImportService.Detect(ctx, request.ToEntity())
}

func (m *mcpImport) discardSource(ctx context.Context, sourceId string) {
	request := dto.NewImportSourceReference(sourceId)
	if err := m.discardImport(ctx, request); err != nil {
		m.errorHandler.Handle(ctx, err, request)
	}
}

func (m *mcpImport) discardImport(ctx context.Context, request dto.ImportSourceReference) error {
	err := validateDiscardImport(&request)
	if err != nil {
		return err
	}
	return m.manageImportService.Discard(ctx, request.SourceId)
}
