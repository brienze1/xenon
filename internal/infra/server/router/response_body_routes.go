package router

import (
	"errors"
	"io"
	"net/http"
	"path"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/infra/server"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

type ResponseBodyRoute struct {
	response adapter.RequestResponseController
}

func NewResponseBodyRoute(response adapter.RequestResponseController) *ResponseBodyRoute {
	return &ResponseBodyRoute{response: response}
}

func (r *router) registerResponseBodyRoutes() wails.Service {
	return wails.NewServiceWithOptions(r.responseBody, wails.ServiceOptions{Route: dto.ResponseBodyRoute})
}

func (r *ResponseBodyRoute) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := r.response.OpenBody(request.Context(), dto.NewTabReference(path.Base(request.URL.Path)))
	if err != nil {
		writeError(writer, err)
		return
	}
	defer func() { _ = body.Content.Close() }()
	writer.Header().Set("Content-Type", body.ContentType)
	writer.WriteHeader(http.StatusOK)
	_, _ = io.Copy(writer, body.Content)
}

func writeError(writer http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var custom errs.BaseError
	if errors.As(err, &custom) {
		status = custom.StatusCode()
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write(server.MarshalError(err))
}
