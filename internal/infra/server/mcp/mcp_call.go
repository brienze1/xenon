package mcp

import (
	"encoding/json"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func newMCPCall(request *sdk.CallToolRequest) dto.MCPCall {
	var session string
	if request.Session != nil {
		session = request.Session.ID()
	}
	var arguments json.RawMessage
	if request.Params != nil {
		arguments = request.Params.Arguments
	}
	return dto.NewMCPCall(session, clientName(request), arguments)
}

func clientName(request *sdk.CallToolRequest) string {
	if info := request.ClientInfo(); info != nil {
		return info.Name
	}
	return ""
}
