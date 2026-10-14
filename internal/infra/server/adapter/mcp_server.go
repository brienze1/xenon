package adapter

import "context"

type MCPServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
