package entity

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func TestNewResponse(t *testing.T) {
	headers := []KeyValue{NewKeyValue("content-type", "application/json"), NewKeyValue("X-A", "1")}
	timings := Timings{Total: time.Second}
	response := NewResponse(200, "OK", "HTTP/1.1", headers, []byte(`{"a":1}`), 3, timings)
	want := Response{
		StatusCode:  200,
		StatusText:  "OK",
		Proto:       "HTTP/1.1",
		Headers:     headers,
		Body:        []byte(`{"a":1}`),
		SizeBytes:   3,
		BodyLength:  7,
		Timings:     timings,
		ContentType: "application/json",
	}
	if !reflect.DeepEqual(response, want) {
		t.Fatalf("NewResponse = %+v, want %+v", response, want)
	}
	if !response.IsOk() || response.Header("X-a") != "1" || response.Header("missing") != "" {
		t.Fatalf("IsOk/Header misbehave: %+v", response)
	}
}

func TestNewFailedResponse(t *testing.T) {
	sendError := NewSendError(senderrorkind.Cancelled, "")
	response := NewFailedResponse(sendError, Timings{Total: time.Millisecond})
	if response.IsOk() || *response.Error != sendError || response.Timings.Total != time.Millisecond {
		t.Fatalf("NewFailedResponse = %+v", response)
	}
	if response.Headers == nil || response.Body == nil || response.HasInlineBody() {
		t.Fatalf("a failed response has empty headers/body and no inline body: %+v", response)
	}
}

func TestResponseBinaryDetection(t *testing.T) {
	cases := []struct {
		contentType string
		body        []byte
		binary      bool
	}{
		{"text/plain", []byte{0, 1, 2}, false},
		{"text/html; charset=utf-8", []byte{0xff}, false},
		{"application/json", []byte{0}, false},
		{"application/problem+xml", []byte{0}, false},
		{"application/javascript", []byte{0}, false},
		{"application/x-www-form-urlencoded", []byte{0}, false},
		{"application/graphql", []byte{0}, false},
		{"application/yaml", []byte{0}, false},
		{"image/png", []byte("\x89PNG\x00"), true},
		{"application/octet-stream", []byte{0, 'a'}, true},
		{"application/octet-stream", []byte{'a', 0}, true},
		{"application/octet-stream", []byte{0xff, 0xfe}, true},
		{"application/octet-stream", []byte("plain ascii"), false},
		{"", []byte("café"), false},
		{"", []byte{}, false},
	}
	for _, tc := range cases {
		response := NewResponse(200, "OK", "HTTP/1.1", []KeyValue{NewKeyValue("Content-Type", tc.contentType)}, tc.body, 0, Timings{})
		if response.Binary != tc.binary {
			t.Errorf("%q %v: binary = %v, want %v", tc.contentType, tc.body, response.Binary, tc.binary)
		}
	}
}

func TestResponseInlineAndMemoryLimits(t *testing.T) {
	cases := []struct {
		size     int
		inline   bool
		inMemory bool
	}{
		{0, true, true},
		{InlineBodyLimit, true, true},
		{InlineBodyLimit + 1, false, true},
		{MemoryBodyLimit, false, true},
		{MemoryBodyLimit + 1, false, false},
	}
	for _, tc := range cases {
		response := NewResponse(200, "OK", "HTTP/1.1", []KeyValue{NewKeyValue("Content-Type", "text/plain")}, bytes.Repeat([]byte("a"), tc.size), 0, Timings{})
		if response.HasInlineBody() != tc.inline || response.KeepsBodyInMemory() != tc.inMemory {
			t.Errorf("size %d: inline %v in memory %v", tc.size, response.HasInlineBody(), response.KeepsBodyInMemory())
		}
	}
	binary := NewResponse(200, "OK", "HTTP/1.1", nil, []byte{0}, 1, Timings{})
	if binary.HasInlineBody() {
		t.Error("a binary body is never inline")
	}
	if InlineBodyLimit != 1048576 || MemoryBodyLimit != 5242880 {
		t.Errorf("limits are 1 MB and 5 MB, got %d and %d", InlineBodyLimit, MemoryBodyLimit)
	}
}

func TestResponseWithoutBodyAndLogValue(t *testing.T) {
	response := NewResponse(201, "Created", "HTTP/1.1", nil, []byte("secret body"), 11, Timings{})
	stripped := response.WithoutBody()
	if stripped.Body != nil || stripped.BodyLength != 11 || response.Body == nil {
		t.Fatalf("WithoutBody must drop only the bytes: %+v", stripped)
	}
	want := map[string]any{"statusCode": 201, "sizeBytes": int64(11), "ok": true}
	if got := response.LogValue(); !reflect.DeepEqual(got, want) {
		t.Fatalf("LogValue = %v, want %v", got, want)
	}
}
