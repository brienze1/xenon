package mcp

import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

func (r *routes) registerSettingsTool(server *sdk.Server) {
	r.addTool(server, r.controllers.Settings)
}
