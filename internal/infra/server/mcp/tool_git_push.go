package mcp

import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

func (r *routes) registerGitPushTool(server *sdk.Server) {
	r.addTool(server, r.controllers.GitPush)
}
