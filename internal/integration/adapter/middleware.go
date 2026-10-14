package adapter

type Middleware interface {
	Handle(ctx MiddlewareContext) error
}
