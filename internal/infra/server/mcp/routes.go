package mcp

import (
	"context"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/middleware"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const endpointPath = "/mcp"

type routes struct {
	controllers    Controllers
	loopbackGuard  adapter.Middleware
	sessionClosing adapter.Middleware
}

func NewRoutes(controllers Controllers, loopbackGuard adapter.Middleware, sessionClosing adapter.Middleware) serverAdapter.MCPRoutes {
	return &routes{controllers: controllers, loopbackGuard: loopbackGuard, sessionClosing: sessionClosing}
}

func (r *routes) Register(server *sdk.Server) {
	r.registerWorkspaceTool(server)
	r.registerTreeTool(server)
	r.registerEnvTool(server)
	r.registerHistoryTool(server)
	r.registerDocsTool(server)
	r.registerResponseTool(server)
	r.registerSnippetTool(server)
	r.registerCookiesTool(server)
	r.registerConsoleTool(server)
	r.registerGitTool(server)
	r.registerScriptsTool(server)
	r.registerSettingsTool(server)
	r.registerWorkspaceEditTool(server)
	r.registerTreeEditTool(server)
	r.registerEnvEditTool(server)
	r.registerDocsEditTool(server)
	r.registerCookiesEditTool(server)
	r.registerGitEditTool(server)
	r.registerSettingsEditTool(server)
	r.registerImportTool(server)
	r.registerExportTool(server)
	r.registerSendTool(server)
	r.registerDeleteTool(server)
	r.registerGitPushTool(server)
}

func (r *routes) Handler(endpoint http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(endpointPath, middleware.HTTPMiddleware(r.sessionClosing, endpoint))
	return middleware.HTTPMiddleware(r.loopbackGuard, mux)
}

func (r *routes) CloseSession(ctx context.Context, sessionId string) {
	_ = r.controllers.Call.CloseSession(ctx, dto.NewMCPSession(sessionId))
}

func (r *routes) addTool(server *sdk.Server, tool adapter.MCPToolController) {
	server.AddTool(newToolDefinition(tool.Tool()), func(ctx context.Context, request *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		return newToolResult(r.controllers.Call.Call(ctx, newMCPCall(request), tool)), nil
	})
}
