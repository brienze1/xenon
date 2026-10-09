package webservice

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/andybalholm/brotli"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	httpClient "github.com/brienze1/xenon/pkg/http"
)

type capturedRequest struct {
	header        http.Header
	host          string
	body          string
	contentLength int64
	chunked       bool
}

type recordingServer struct {
	server   *httptest.Server
	requests []capturedRequest
	mu       sync.Mutex
}

func newRecordingServer(t *testing.T, handler http.HandlerFunc) *recordingServer {
	recorder := &recordingServer{}
	recorder.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		recorder.mu.Lock()
		recorder.requests = append(recorder.requests, capturedRequest{
			header:        r.Header.Clone(),
			host:          r.Host,
			body:          string(body),
			contentLength: r.ContentLength,
			chunked:       len(r.TransferEncoding) > 0,
		})
		recorder.mu.Unlock()
		if handler != nil {
			handler(w, r)
		}
	}))
	t.Cleanup(recorder.server.Close)
	return recorder
}

func (r *recordingServer) last(t *testing.T) capturedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.requests) == 0 {
		t.Fatal("the server received no request")
	}
	return r.requests[len(r.requests)-1]
}

func send(request entity.EffectiveRequest) entity.Response {
	return NewHttpWebservice(httpClient.NewClient).Send(context.Background(), request)
}

func getRequest(url string) entity.EffectiveRequest {
	return entity.EffectiveRequest{Method: http.MethodGet, URL: url, Settings: entity.DefaultRequestSettings()}
}

func TestSendReturnsTheResponseWithTimings(t *testing.T) {
	server := newRecordingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.Header().Add("Set-Cookie", "b=2")
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusAccepted)
		w.(http.Flusher).Flush()
		time.Sleep(20 * time.Millisecond)
		_, _ = w.Write([]byte("hello"))
	})
	url := strings.Replace(server.server.URL, "127.0.0.1", "localhost", 1)
	response := send(getRequest(url + "/a"))
	if !response.IsOk() || response.StatusCode != http.StatusAccepted || response.StatusText != "Accepted" || response.Proto != "HTTP/1.1" {
		t.Fatalf("response = %+v", response)
	}
	if string(response.Body) != "hello" || response.SizeBytes != 5 || response.ContentType != "text/plain" {
		t.Fatalf("body = %q size %d type %q", response.Body, response.SizeBytes, response.ContentType)
	}
	if response.Header("Set-Cookie") != "b=2" {
		t.Fatalf("headers keep the received order of repeated values: %+v", response.Headers)
	}
	timings := response.Timings
	if timings.Dns <= 0 || timings.Connect <= 0 || timings.FirstByte < 15*time.Millisecond || timings.Download < 15*time.Millisecond || timings.Total < timings.FirstByte+timings.Download {
		t.Fatalf("timings = %+v", timings)
	}
	if timings.Tls != 0 {
		t.Fatalf("plain HTTP has no TLS time: %+v", timings)
	}
}

func TestSendMeasuresTheTlsHandshake(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	request := getRequest(server.URL)
	request.Settings.VerifySsl = false
	response := send(request)
	if !response.IsOk() || response.Timings.Tls <= 0 {
		t.Fatalf("response = %+v", response)
	}
}

func TestSendAppliesHeaders(t *testing.T) {
	server := newRecordingServer(t, nil)
	request := entity.EffectiveRequest{
		Method:   http.MethodPost,
		URL:      server.server.URL,
		Settings: entity.DefaultRequestSettings(),
		Headers:  []entity.KeyValue{entity.NewKeyValue("host", "virtual.test"), entity.NewKeyValue("X-Multi", "1"), entity.NewKeyValue("X-Multi", "2")},
		Body:     entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: "payload"}},
		AutoHeaders: []entity.AutoHeader{
			entity.NewAutoHeader("Content-Type", "application/json"),
			entity.NewComputedAutoHeader("Content-Length", entity.ComputedAtSend),
			{Key: "Accept", Value: "*/*", Disabled: true},
		},
	}
	if response := send(request); !response.IsOk() {
		t.Fatalf("response = %+v", response)
	}
	received := server.last(t)
	if received.host != "virtual.test" || received.header.Get("Content-Type") != "application/json" || received.body != "payload" {
		t.Fatalf("received = %+v", received)
	}
	if strings.Join(received.header.Values("X-Multi"), ",") != "1,2" {
		t.Fatalf("repeated user headers are all sent: %v", received.header.Values("X-Multi"))
	}
	if _, found := received.header["User-Agent"]; found {
		t.Fatalf("no User-Agent unless Xenon sends one: %v", received.header)
	}
	if _, found := received.header["Accept"]; found {
		t.Fatal("a disabled auto header is not sent")
	}
	if received.contentLength != 7 || received.chunked {
		t.Fatalf("content length = %d chunked %v", received.contentLength, received.chunked)
	}
}

func TestSendWithContentLengthSwitchedOffIsChunked(t *testing.T) {
	server := newRecordingServer(t, nil)
	request := entity.EffectiveRequest{
		Method:              http.MethodPost,
		URL:                 server.server.URL,
		Settings:            entity.DefaultRequestSettings(),
		Body:                entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: "payload"}},
		DisabledAutoHeaders: []string{"Content-Length"},
	}
	send(request)
	received := server.last(t)
	if !received.chunked || received.contentLength != -1 || received.body != "payload" {
		t.Fatalf("received = %+v", received)
	}
}

func TestSendFormDataUsesTheWriterContentTypeOnlyForTheAutoHeader(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "note.txt")
	if err := os.WriteFile(path, []byte("file body"), 0o600); err != nil {
		t.Fatal(err)
	}
	body := entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{
		{Key: "note", Type: formdatatype.Text, Value: "hello", Enabled: true},
		{Key: "off", Type: formdatatype.Text, Value: "no"},
		{Key: "file", Type: formdatatype.File, Src: entity.FileRef{Src: path}, Enabled: true},
	}}
	server := newRecordingServer(t, func(w http.ResponseWriter, r *http.Request) {})
	automatic := entity.EffectiveRequest{Method: http.MethodPost, URL: server.server.URL, Settings: entity.DefaultRequestSettings(), Body: body,
		AutoHeaders: []entity.AutoHeader{entity.NewComputedAutoHeader("Content-Type", "multipart/form-data; boundary=<calculated when request is sent>")}}
	send(automatic)
	received := server.last(t)
	parsed, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(received.body))
	parsed.Header = received.header
	if err := parsed.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("multipart body: %v (%s)", err, received.header.Get("Content-Type"))
	}
	file, header, _ := parsed.FormFile("file")
	content, _ := io.ReadAll(file)
	if parsed.FormValue("note") != "hello" || parsed.FormValue("off") != "" || header.Filename != "note.txt" || string(content) != "file body" {
		t.Fatalf("form = %+v", parsed.MultipartForm)
	}

	own := automatic
	own.Headers = []entity.KeyValue{entity.NewKeyValue("Content-Type", "multipart/mixed")}
	own.AutoHeaders = nil
	send(own)
	if got := server.last(t).header.Get("Content-Type"); got != "multipart/mixed" {
		t.Fatalf("a user Content-Type is kept, got %q", got)
	}
}

func TestSendEncodesUrlEncodedAndGraphQlBodies(t *testing.T) {
	server := newRecordingServer(t, nil)
	request := getRequest(server.server.URL)
	request.Method = http.MethodPost
	request.Body = entity.Body{Mode: bodymode.UrlEncoded, UrlEncoded: []entity.KeyValue{entity.NewKeyValue("a b", "c&d"), {Key: "off", Value: "1"}, entity.NewKeyValue("e", "f")}}
	send(request)
	if got := server.last(t).body; got != "a+b=c%26d&e=f" {
		t.Fatalf("urlencoded = %q", got)
	}
	for variables, want := range map[string]string{
		` { "a" : 1 } `: `{"query":"{ me }","variables":{"a":1}}`,
		"":              `{"query":"{ me }","variables":{}}`,
		"not json":      `{"query":"{ me }","variables":{}}`,
	} {
		request.Body = entity.Body{Mode: bodymode.GraphQl, GraphQl: entity.BodyGraphQl{Query: "{ me }", Variables: variables}}
		send(request)
		if got := server.last(t).body; got != want {
			t.Errorf("graphql %q = %q, want %q", variables, got, want)
		}
	}
	request.Body = entity.Body{Mode: bodymode.None}
	send(request)
	if received := server.last(t); received.body != "" || received.contentLength != 0 {
		t.Fatalf("no body = %+v", received)
	}
}

func TestSendFileErrors(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "missing.bin")
	server := newRecordingServer(t, nil)
	cases := []struct {
		body entity.Body
		kind string
	}{
		{entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: missing}}, "missing_file"},
		{entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: directory}}, "other"},
		{entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{{Key: "f", Type: formdatatype.File, Src: entity.FileRef{Src: missing}, Enabled: true}}}, "missing_file"},
	}
	for _, tc := range cases {
		request := getRequest(server.server.URL)
		request.Body = tc.body
		response := send(request)
		if response.IsOk() || response.Error.Kind.String() != tc.kind {
			t.Errorf("%s: response = %+v", tc.body.Mode, response)
		}
	}
	if len(server.requests) != 0 {
		t.Fatal("nothing is sent when a body file cannot be read")
	}
	request := getRequest(server.server.URL)
	request.Body = entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: missing}}
	if got := send(request).Error.Message; got != "File not found: "+missing {
		t.Fatalf("message = %q", got)
	}
}

func TestSendRejectsInvalidTargets(t *testing.T) {
	cases := map[string]string{
		"":                   "Enter a URL to send the request",
		"   ":                "Enter a URL to send the request",
		"http://[::1/users":  "Invalid URL: http://[::1/users",
		"ftp://example.test": "Invalid URL: ftp://example.test",
		"http://":            "Invalid URL: http://",
		"/relative":          "Invalid URL: /relative",
	}
	for url, message := range cases {
		response := send(getRequest(url))
		if response.IsOk() || response.Error.Kind != senderrorkind.InvalidUrl || response.Error.Message != message {
			t.Errorf("%q: %+v", url, response.Error)
		}
	}
	request := getRequest("HTTPS://example.test")
	request.Method = "BAD METHOD"
	if response := send(request); response.Error == nil || response.Error.Kind != senderrorkind.Other {
		t.Fatalf("an invalid method fails as other: %+v", response.Error)
	}
}

func TestSendFailuresKeepTimings(t *testing.T) {
	listener := httptest.NewServer(nil)
	url := listener.URL
	listener.Close()
	response := send(getRequest(url))
	if response.IsOk() || response.Error.Kind != senderrorkind.ConnectionRefused || response.Error.Message != "Could not connect to "+strings.TrimPrefix(url, "http://")+" (connection refused)" {
		t.Fatalf("response = %+v", response.Error)
	}
	if response.Timings.Total <= 0 || response.Timings.Download != 0 || response.Timings.FirstByte != 0 {
		t.Fatalf("timings = %+v", response.Timings)
	}
}

func TestSendTimeoutAndCancel(t *testing.T) {
	server := newRecordingServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})
	request := getRequest(server.server.URL)
	request.Settings.TimeoutMs = 50
	if response := send(request); response.Error == nil || response.Error.Kind != senderrorkind.Timeout || response.Error.Message != "Timed out after 50 ms" {
		t.Fatalf("timeout = %+v", response.Error)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	request.Settings.TimeoutMs = 0
	if response := NewHttpWebservice(httpClient.NewClient).Send(ctx, request); response.Error == nil || response.Error.Kind != senderrorkind.Timeout {
		t.Fatalf("deadline = %+v", response.Error)
	}
	cancelled, cancelNow := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancelNow)
	if response := NewHttpWebservice(httpClient.NewClient).Send(cancelled, request); response.Error == nil || response.Error.Kind != senderrorkind.Cancelled {
		t.Fatalf("cancel = %+v", response.Error)
	}
}

func TestSendBodyReadFailure(t *testing.T) {
	server := newRecordingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("short"))
		conn, _, _ := w.(http.Hijacker).Hijack()
		_ = conn.Close()
	})
	response := send(getRequest(server.server.URL))
	if response.IsOk() || response.Error.Kind != senderrorkind.Other {
		t.Fatalf("a broken body fails: %+v", response.Error)
	}
}

func TestDecode(t *testing.T) {
	plain := []byte("hello, xenon")
	var gzipped, zlibbed, deflated, brotlied bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipped)
	_, _ = gzipWriter.Write(plain)
	_ = gzipWriter.Close()
	zlibWriter := zlib.NewWriter(&zlibbed)
	_, _ = zlibWriter.Write(plain)
	_ = zlibWriter.Close()
	flateWriter, _ := flate.NewWriter(&deflated, flate.DefaultCompression)
	_, _ = flateWriter.Write(plain)
	_ = flateWriter.Close()
	brotliWriter := brotli.NewWriter(&brotlied)
	_, _ = brotliWriter.Write(plain)
	_ = brotliWriter.Close()
	garbage := []byte("not compressed at all")
	cases := []struct {
		encoding string
		raw      []byte
		want     []byte
	}{
		{" GZIP ", gzipped.Bytes(), plain},
		{"deflate", zlibbed.Bytes(), plain},
		{"deflate", deflated.Bytes(), plain},
		{"br", brotlied.Bytes(), plain},
		{"", plain, plain},
		{"identity", plain, plain},
		{"gzip", garbage, garbage},
		{"gzip", gzipped.Bytes()[:gzipped.Len()-4], gzipped.Bytes()[:gzipped.Len()-4]},
		{"deflate", garbage, garbage},
		{"br", garbage, garbage},
	}
	for _, tc := range cases {
		if got := decode(tc.encoding, tc.raw); !bytes.Equal(got, tc.want) {
			t.Errorf("decode(%q) = %q, want %q", tc.encoding, got, tc.want)
		}
	}
}

func TestSendDecodesAndKeepsTheWireSize(t *testing.T) {
	var gzipped bytes.Buffer
	writer := gzip.NewWriter(&gzipped)
	_, _ = writer.Write([]byte("hello, xenon"))
	_ = writer.Close()
	server := newRecordingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(gzipped.Bytes())
	})
	request := getRequest(server.server.URL)
	request.AutoHeaders = []entity.AutoHeader{entity.NewAutoHeader("Accept-Encoding", "gzip, deflate, br")}
	response := send(request)
	if string(response.Body) != "hello, xenon" || response.SizeBytes != int64(gzipped.Len()) || response.Header("Content-Encoding") != "gzip" {
		t.Fatalf("response = %q %d %+v", response.Body, response.SizeBytes, response.Headers)
	}
	if got := server.last(t).header.Get("Accept-Encoding"); got != "gzip, deflate, br" {
		t.Fatalf("accept-encoding = %q", got)
	}
}
