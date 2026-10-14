package mcp

import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

func (r *routes) registerTreeEditTool(server *sdk.Server) {
	r.addTool(server, r.controllers.TreeEdit)
}
