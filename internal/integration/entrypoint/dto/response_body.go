package dto

import (
	"io"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ResponseBody struct {
	Content     io.ReadCloser
	ContentType string
}

func NewResponseBody(response entity.Response, content io.ReadCloser) ResponseBody {
	return ResponseBody{Content: content, ContentType: response.ContentType}
}
