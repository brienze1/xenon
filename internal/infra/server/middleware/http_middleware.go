package middleware

import (
	"net/http"

	"github.com/brienze1/xenon/internal/infra/server/context"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

func HTTPMiddleware(middleware adapter.Middleware, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := middleware.Handle(context.HTTP(writer, request, next)); err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
