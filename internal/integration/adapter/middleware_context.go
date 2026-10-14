package adapter

import "context"

type MiddlewareContext interface {
	Context() context.Context
	Host() string
	Port() int
	Header(key string) (string, bool)
	Method() string
	Next() error
	Forbid() error
}
