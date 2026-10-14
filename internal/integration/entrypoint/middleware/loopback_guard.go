package middleware

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	rejectedByHost   = "host"
	rejectedByOrigin = "origin"
	originHeader     = "Origin"
)

type loopbackGuard struct{}

func NewLoopbackGuardMiddleware() adapter.Middleware {
	return &loopbackGuard{}
}

func (l *loopbackGuard) Handle(ctx adapter.MiddlewareContext) error {
	if !entity.IsAILoopbackHost(ctx.Host(), ctx.Port()) {
		return l.reject(ctx, rejectedByHost)
	}
	if origin, sent := ctx.Header(originHeader); sent && !entity.IsAILoopbackOrigin(origin, ctx.Port()) {
		return l.reject(ctx, rejectedByOrigin)
	}
	return ctx.Next()
}

func (l *loopbackGuard) reject(ctx adapter.MiddlewareContext, reason string) error {
	origin, _ := ctx.Header(originHeader)
	logger.Warn(ctx.Context(), nil, "MCP request rejected", map[string]any{"reason": reason, "host": ctx.Host(), "origin": origin, "method": ctx.Method()})
	return ctx.Forbid()
}
