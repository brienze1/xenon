package dto

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type HttpResponse struct {
	Header     http.Header
	Status     string
	Proto      string
	Body       []byte
	StatusCode int
	WireSize   int64
}

func NewHttpResponse(response *http.Response, body []byte, wireSize int64) HttpResponse {
	return HttpResponse{
		Header:     response.Header,
		Status:     response.Status,
		Proto:      response.Proto,
		Body:       body,
		StatusCode: response.StatusCode,
		WireSize:   wireSize,
	}
}

func (h HttpResponse) ToEntity(timings entity.Timings) entity.Response {
	return entity.NewResponse(h.StatusCode, h.statusText(), h.Proto, h.headers(), h.Body, h.WireSize, timings)
}

func (h HttpResponse) statusText() string {
	return strings.TrimPrefix(h.Status, strconv.Itoa(h.StatusCode)+" ")
}

func (h HttpResponse) headers() []entity.KeyValue {
	keys := make([]string, 0, len(h.Header))
	for key := range h.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	headers := make([]entity.KeyValue, 0, len(keys))
	for _, key := range keys {
		for _, value := range h.Header[key] {
			headers = append(headers, entity.NewKeyValue(key, value))
		}
	}
	return headers
}
