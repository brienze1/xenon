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
	mcpScriptsCheckAction     = "check"
	mcpScriptsReferenceAction = "reference"
	mcpScriptsSnippetsAction  = "snippets"
)

var scriptsAITool = entity.AITool{
	Name:            "scripts",
	Description:     "Check a script for errors or read the scripting reference.",
	Actions:         []string{mcpScriptsCheckAction, mcpScriptsReferenceAction, mcpScriptsSnippetsAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpScripts struct {
	errorHandler       adapter.ErrorHandler
	scriptToolsService appAdapter.ScriptToolsService
}

func NewMCPScripts(errorHandler adapter.ErrorHandler, scriptToolsService appAdapter.ScriptToolsService) adapter.MCPToolController {
	return &mcpScripts{errorHandler: errorHandler, scriptToolsService: scriptToolsService}
}

func (m *mcpScripts) Tool() dto.MCPTool {
	return dto.NewMCPTool(scriptsAITool)
}

func (m *mcpScripts) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpScriptsCheckAction:
		result, err = m.check(ctx, call)
	case mcpScriptsReferenceAction:
		result, err = m.reference(ctx)
	case mcpScriptsSnippetsAction:
		result, err = m.snippets(ctx)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpScripts) check(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPScriptCheck
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.CheckScript()
	err = validateCheckScript(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	return dto.NewMCPToolResult(dto.NewScriptCheck(m.scriptToolsService.Check(ctx, request.Text))), nil
}

func (m *mcpScripts) reference(ctx context.Context) (dto.MCPToolResult, error) {
	members, err := m.scriptToolsService.Reference(ctx)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPScriptReference(dto.NewScriptReference(members))), nil
}

func (m *mcpScripts) snippets(ctx context.Context) (dto.MCPToolResult, error) {
	snippets, err := m.scriptToolsService.Snippets(ctx)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPList(dto.NewMCPScriptSnippets(dto.NewScriptSnippets(snippets)))), nil
}
