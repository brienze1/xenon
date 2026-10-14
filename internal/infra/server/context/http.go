package context

import (
	"context"
	"net"
	"net/http"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type httpCtx struct {
	writer  http.ResponseWriter
	request *http.Request
	next    http.Handler
}

func HTTP(writer http.ResponseWriter, request *http.Request, next http.Handler) adapter.MiddlewareContext {
	return &httpCtx{writer: writer, request: request, next: next}
}

func (h *httpCtx) Context() context.Context {
	return h.request.Context()
}

func (h *httpCtx) Host() string {
	return h.request.Host
}

func (h *httpCtx) Port() int {
	if address, ok := h.request.Context().Value(http.LocalAddrContextKey).(*net.TCPAddr); ok {
		return address.Port
	}
	return 0
}

func (h *httpCtx) Header(key string) (string, bool) {
	values := h.request.Header.Values(key)
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func (h *httpCtx) Method() string {
	return h.request.Method
}

func (h *httpCtx) Next() error {
	h.next.ServeHTTP(h.writer, h.request)
	return nil
}

func (h *httpCtx) Forbid() error {
	http.Error(h.writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	return nil
}
