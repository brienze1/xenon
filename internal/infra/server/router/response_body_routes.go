package router

import (
	"errors"
	"io"
	"net/http"
	"path"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

type ResponseBodyRoute struct {
	response adapter.RequestResponseController
}

func NewResponseBodyRoute() *ResponseBodyRoute {
	return &ResponseBodyRoute{response: controller.GetRequestResponseController()}
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
