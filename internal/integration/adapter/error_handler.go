package adapter

import "context"

type ErrorHandler interface {
	Handle(ctx context.Context, err error, event any) error
}
