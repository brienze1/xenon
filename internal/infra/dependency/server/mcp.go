package server

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/infra/server/mcp"
)

var mcpServerOnce sync.Once
var mcpServerInstance *mcp.Server

func GetMCPServer() *mcp.Server {
	if mcpServerInstance == nil {
		mcpServerOnce.Do(func() {
			mcpServerInstance = mcp.NewServer(integration.GetAIEventPublisher())
		})
	}
	return mcpServerInstance
}
