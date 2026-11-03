package adapter

import (
	"context"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPRoutes interface {
	Register(server *sdk.Server)
	Handler(endpoint http.Handler) http.Handler
	OpenSession(ctx context.Context, sessionId string)
	CloseSession(ctx context.Context, sessionId string)
}
