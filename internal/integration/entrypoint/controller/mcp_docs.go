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
	mcpDocsGetAction  = "get"
	mcpDocsPlanAction = "plan"
)

var docsAITool = entity.AITool{
	Name:            "docs",
	Description:     "Read a collection's OpenAPI document or preview the plan for an edited one.",
	Actions:         []string{mcpDocsGetAction, mcpDocsPlanAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpDocs struct {
	errorHandler               adapter.ErrorHandler
	searchTreeService          appAdapter.SearchTreeService
	manageDocumentationService appAdapter.ManageDocumentationService
}

func NewMCPDocs(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	manageDocumentationService appAdapter.ManageDocumentationService,
) adapter.MCPToolController {
	return &mcpDocs{
		errorHandler:               errorHandler,
		searchTreeService:          searchTreeService,
		manageDocumentationService: manageDocumentationService,
	}
}

func (m *mcpDocs) Tool() dto.MCPTool {
	return dto.NewMCPTool(docsAITool)
}

func (m *mcpDocs) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpDocsGetAction:
		result, err = m.get(ctx, call)
	case mcpDocsPlanAction:
		result, err = m.plan(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpDocs) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPDocsInput
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	collection, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.CollectionItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.DocumentationReference(call.LocalId)
	err = validateGetDocumentation(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	documentation, err := m.manageDocumentationService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPDocs(dto.NewDocumentation(*documentation))).WithTarget(collection.Target()), nil
}

func (m *mcpDocs) plan(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPDocsInput
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	collection, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.CollectionItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.PlanInput(call.LocalId)
	err = validatePlanDocumentation(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	plan, err := m.manageDocumentationService.Plan(ctx, request.LocalId, request.Id, request.Yaml)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPDocsPlan(dto.NewDocumentationPlan(*plan))).WithTarget(collection.Target()), nil
}
