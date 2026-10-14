package external

import (
	"sync"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/mcp"
)

var aiAccessListenerOnce sync.Once
var aiAccessListenerInstance serverAdapter.MCPListener

func GetAIAccessListener() serverAdapter.MCPListener {
	if aiAccessListenerInstance == nil {
		aiAccessListenerOnce.Do(func() {
			aiAccessListenerInstance = mcp.NewListener()
		})
	}
	return aiAccessListenerInstance
}

func SetAIAccessListener(aiAccessListener serverAdapter.MCPListener) {
	aiAccessListenerOnce.Do(func() {})
	aiAccessListenerInstance = aiAccessListener
}
