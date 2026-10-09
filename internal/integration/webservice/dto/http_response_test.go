package dto

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestHttpResponseToEntity(t *testing.T) {
	raw := &http.Response{
		StatusCode: 200,
		Status:     "200 All Good",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"X-B": {"2"}, "Content-Type": {"text/plain"}, "Set-Cookie": {"b=2", "a=1"}},
	}
	timings := entity.Timings{Total: time.Second}
	response := NewHttpResponse(raw, []byte("hello"), 3).ToEntity(timings)
	if response.StatusText != "All Good" || response.StatusCode != 200 || response.Proto != "HTTP/1.1" || string(response.Body) != "hello" || response.SizeBytes != 3 || response.Timings != timings {
		t.Fatalf("response = %+v", response)
	}
	want := []entity.KeyValue{
		entity.NewKeyValue("Content-Type", "text/plain"),
		entity.NewKeyValue("Set-Cookie", "b=2"),
		entity.NewKeyValue("Set-Cookie", "a=1"),
		entity.NewKeyValue("X-B", "2"),
	}
	if !reflect.DeepEqual(response.Headers, want) {
		t.Fatalf("headers = %+v", response.Headers)
	}
}
