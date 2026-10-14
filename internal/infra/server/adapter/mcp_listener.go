package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type MCPListener interface {
	adapter.AIAccessListener
	Route(routes MCPRoutes)
	Watch(clientsChanged func(ctx context.Context))
}
