package server

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/external"
	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/mcp"
)

var mcpServerOnce sync.Once
var mcpServerInstance serverAdapter.MCPServer

func GetMCPServer() serverAdapter.MCPServer {
	if mcpServerInstance == nil {
		mcpServerOnce.Do(func() {
			mcpServerInstance = mcp.NewServer(controller.GetAIAccessController(), external.GetAIAccessListener(), GetMCPRoutes())
		})
	}
	return mcpServerInstance
}
