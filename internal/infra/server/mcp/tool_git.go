package mcp

import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

func (r *routes) registerGitTool(server *sdk.Server) {
	r.addTool(server, r.controllers.Git)
}
