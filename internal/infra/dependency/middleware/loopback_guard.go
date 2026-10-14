package middleware

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	entrypointMiddleware "github.com/brienze1/xenon/internal/integration/entrypoint/middleware"
)

var loopbackGuardMiddlewareOnce sync.Once
var loopbackGuardMiddlewareInstance adapter.Middleware

func GetLoopbackGuardMiddleware() adapter.Middleware {
	if loopbackGuardMiddlewareInstance == nil {
		loopbackGuardMiddlewareOnce.Do(func() {
			loopbackGuardMiddlewareInstance = entrypointMiddleware.NewLoopbackGuardMiddleware()
		})
	}
	return loopbackGuardMiddlewareInstance
}
