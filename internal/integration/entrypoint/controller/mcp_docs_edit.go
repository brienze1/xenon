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

const mcpDocsEditApplyAction = "apply"

var docsEditAITool = entity.AITool{
	Name:            "docs_edit",
	Description:     "Apply an edited OpenAPI document to a collection.",
	Actions:         []string{mcpDocsEditApplyAction},
	Effect:          aitooleffect.Destructive,
	WorkspaceScoped: true,
}

type mcpDocsEdit struct {
	errorHandler               adapter.ErrorHandler
	searchTreeService          appAdapter.SearchTreeService
	manageDocumentationService appAdapter.ManageDocumentationService
	autoSyncService            appAdapter.AutoSyncService
}

func NewMCPDocsEdit(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	manageDocumentationService appAdapter.ManageDocumentationService,
	autoSyncService appAdapter.AutoSyncService,
) adapter.MCPToolController {
	return &mcpDocsEdit{
		errorHandler:               errorHandler,
		searchTreeService:          searchTreeService,
		manageDocumentationService: manageDocumentationService,
		autoSyncService:            autoSyncService,
	}
}

func (m *mcpDocsEdit) Tool() dto.MCPTool {
	return dto.NewMCPTool(docsEditAITool, dto.NewMCPToolArguments[dto.MCPDocsInput](mcpDocsEditApplyAction))
}

func (m *mcpDocsEdit) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpDocsEditApplyAction:
		result, err = m.apply(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpDocsEdit) apply(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPDocsInput
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	collection, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.CollectionItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.ApplyInput(call.LocalId)
	err = validateApplyDocumentation(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	applied, err := m.manageDocumentationService.Apply(ctx, request.LocalId, request.Id, request.Yaml, request.ConfirmDeletions, request.KeepDeleted)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Saved, request.LocalId)
	plan := dto.NewDocumentationApplied(*applied).DocumentationPlan
	return dto.NewMCPToolResult(dto.NewMCPDocsPlan(plan)).WithTarget(collection.Target()).WithChanges(aichangekind.Tree), nil
}
