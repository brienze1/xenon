package adapter

import "context"

type Consumer interface {
	Register(ctx context.Context)
}
