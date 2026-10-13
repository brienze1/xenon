package mcp

import (
	"net/http"
	"strconv"

	"github.com/brienze1/xenon/pkg/logger"
)

const (
	loopbackIP       = "127.0.0.1"
	loopbackName     = "localhost"
	loopbackScheme   = "http://"
	rejectedByHost   = "host"
	rejectedByOrigin = "origin"
	originHeader     = "Origin"
)

func loopbackOnly(port int, next http.Handler) http.Handler {
	suffix := ":" + strconv.Itoa(port)
	hosts := map[string]bool{loopbackIP + suffix: true, loopbackName + suffix: true}
	origins := map[string]bool{
		loopbackScheme + loopbackIP:            true,
		loopbackScheme + loopbackIP + suffix:   true,
		loopbackScheme + loopbackName:          true,
		loopbackScheme + loopbackName + suffix: true,
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !hosts[request.Host] {
			reject(writer, request, rejectedByHost)
			return
		}
		if origin, sent := request.Header[originHeader]; sent && !origins[origin[0]] {
			reject(writer, request, rejectedByOrigin)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func reject(writer http.ResponseWriter, request *http.Request, reason string) {
	logger.Warn(request.Context(), nil, "MCP request rejected", map[string]any{"reason": reason, "host": request.Host, "origin": request.Header.Get(originHeader), "method": request.Method})
	http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}
