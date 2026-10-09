package route

import (
	"io"
	"net/http"
	"path"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ResponseBodyRoute struct {
	response adapter.RequestResponseController
}

func NewResponseBodyRoute() *ResponseBodyRoute {
	return &ResponseBodyRoute{response: controller.GetRequestResponse()}
}

func (r *ResponseBodyRoute) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := r.response.OpenBody(request.Context(), dto.TabReference{TabId: path.Base(request.URL.Path)})
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer func() { _ = body.Content.Close() }()
	writer.Header().Set("Content-Type", body.ContentType)
	writer.WriteHeader(http.StatusOK)
	_, _ = io.Copy(writer, body.Content)
}
