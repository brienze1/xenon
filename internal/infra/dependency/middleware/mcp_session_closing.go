package middleware

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	entrypointMiddleware "github.com/brienze1/xenon/internal/integration/entrypoint/middleware"
)

var mcpSessionClosingMiddlewareOnce sync.Once
var mcpSessionClosingMiddlewareInstance adapter.Middleware

func GetMCPSessionClosingMiddleware() adapter.Middleware {
	if mcpSessionClosingMiddlewareInstance == nil {
		mcpSessionClosingMiddlewareOnce.Do(func() {
			mcpSessionClosingMiddlewareInstance = entrypointMiddleware.NewMCPSessionClosingMiddleware(controller.GetMCPCallController())
		})
	}
	return mcpSessionClosingMiddlewareInstance
}
