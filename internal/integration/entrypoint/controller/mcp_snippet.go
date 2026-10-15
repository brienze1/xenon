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
	mcpSnippetLanguagesAction = "languages"
	mcpSnippetGenerateAction  = "generate"
	mcpSnippetSecretsAction   = "Generating a snippet with secrets"
)

var snippetAITool = entity.AITool{
	Name:            "snippet",
	Description:     "Generate a code snippet for a request (secrets masked).",
	Actions:         []string{mcpSnippetLanguagesAction, mcpSnippetGenerateAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpSnippet struct {
	errorHandler           adapter.ErrorHandler
	generateSnippetService appAdapter.GenerateSnippetService
	searchTreeService      appAdapter.SearchTreeService
	saveRequestService     appAdapter.SaveRequestService
}

func NewMCPSnippet(
	errorHandler adapter.ErrorHandler,
	generateSnippetService appAdapter.GenerateSnippetService,
	searchTreeService appAdapter.SearchTreeService,
	saveRequestService appAdapter.SaveRequestService,
) adapter.MCPToolController {
	return &mcpSnippet{
		errorHandler:           errorHandler,
		generateSnippetService: generateSnippetService,
		searchTreeService:      searchTreeService,
		saveRequestService:     saveRequestService,
	}
}

func (m *mcpSnippet) Tool() dto.MCPTool {
	return dto.NewMCPTool(snippetAITool, dto.NewMCPToolArguments[dto.MCPSnippetInput](mcpSnippetGenerateAction))
}

func (m *mcpSnippet) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpSnippetLanguagesAction:
		result = m.languages(ctx)
	case mcpSnippetGenerateAction:
		result, err = m.generate(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpSnippet) languages(ctx context.Context) dto.MCPToolResult {
	languages := dto.NewSnippetLanguageList(m.generateSnippetService.Languages(ctx))
	return dto.NewMCPToolResult(dto.NewMCPList(dto.NewMCPSnippetLanguages(languages)))
}

func (m *mcpSnippet) generate(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPSnippetInput
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if arguments.RequestsSecrets() {
		return dto.MCPToolResult{}, mcpErrors.RejectHumanOnlyAIAction(mcpSnippetSecretsAction)
	}

	resolved, target, err := resolveMCPSendTarget(ctx, m.searchTreeService, m.saveRequestService, call, arguments.SendTarget())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	request := arguments.GenerateSnippet(call.LocalId, call.TabId(), resolved)
	err = validateGenerateSnippet(&request)
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	generated, err := m.generateSnippetService.Generate(ctx, request.ToEntity())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	return dto.NewMCPToolResult(dto.NewMCPSnippet(dto.NewSnippet(*generated))).WithTarget(target), nil
}
