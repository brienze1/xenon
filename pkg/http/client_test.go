package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientFollowsRedirectsOnlyWhenAsked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			http.Redirect(w, r, "/new", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	for _, follow := range []bool{true, false} {
		client := HttpClient(Options{FollowRedirects: follow, VerifySsl: true}, &recordingLogger{})
		request, _ := http.NewRequest(http.MethodGet, server.URL+"/old", nil)
		response, err := client.Do(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		client.CloseIdleConnections()
		want := http.StatusFound
		if follow {
			want = http.StatusNoContent
		}
		if response.StatusCode != want {
			t.Fatalf("follow=%v: status %d, want %d", follow, response.StatusCode, want)
		}
	}
}

func TestClientVerifiesCertificatesOnlyWhenAsked(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer server.Close()

	for _, verify := range []bool{true, false} {
		client := HttpClient(Options{VerifySsl: verify, Timeout: time.Second}, &recordingLogger{})
		request, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		response, err := client.Do(context.Background(), request)
		if verify != (err != nil) {
			t.Fatalf("verify=%v: err = %v", verify, err)
		}
		if response != nil {
			_ = response.Body.Close()
		}
	}
}

type logEntry struct {
	err      error
	message  string
	metadata map[string]any
}

type recordingLogger struct {
	debug []logEntry
	error []logEntry
}

func (r *recordingLogger) Debug(_ context.Context, message string, metadata ...any) {
	r.debug = append(r.debug, logEntry{message: message, metadata: metadata[0].(map[string]any)})
}

func (r *recordingLogger) Error(_ context.Context, err error, message string, metadata ...any) {
	r.error = append(r.error, logEntry{err: err, message: message, metadata: metadata[0].(map[string]any)})
}

func TestClientLogsSuccessAtDebugWithMaskedSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Set-Cookie", "session=secret")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	logger := &recordingLogger{}
	client := HttpClient(Options{VerifySsl: true, Timeout: time.Second}, logger)
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/path?access_token=abc&page=2&Api_Key=k", nil)
	request.Header.Set("Authorization", "Bearer abc")
	request.Header.Set("Cookie", "a=b")
	request.Header.Set("Proxy-Authorization", "Basic x")
	request.Header.Set("X-Api-Key", "key")
	request.Header.Set("Accept", "text/plain")
	response, err := client.Do(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if len(logger.debug) != 1 || len(logger.error) != 0 {
		t.Fatalf("debug=%d error=%d, want one debug entry", len(logger.debug), len(logger.error))
	}
	metadata := logger.debug[0].metadata
	if metadata["method"] != http.MethodGet || metadata["status"] != http.StatusOK {
		t.Fatalf("metadata = %v", metadata)
	}
	if got := metadata["url"]; got != server.URL+"/path?access_token=******&page=2&Api_Key=******" {
		t.Fatalf("url = %v", got)
	}
	headers := metadata["headers"].(map[string][]string)
	for _, key := range []string{"Authorization", "Cookie", "Proxy-Authorization", "X-Api-Key"} {
		if headers[key][0] != maskedValue {
			t.Fatalf("%s = %v, want masked", key, headers[key])
		}
	}
	if headers["Accept"][0] != "text/plain" {
		t.Fatalf("Accept = %v", headers["Accept"])
	}
	if metadata["response_headers"].(map[string][]string)["Set-Cookie"][0] != maskedValue {
		t.Fatal("Set-Cookie must be masked")
	}
	if _, found := metadata["duration_ms"]; !found {
		t.Fatal("duration must be logged")
	}
}

func TestClientLogsErrorStatusAndFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	logger := &recordingLogger{}
	client := HttpClient(Options{VerifySsl: true, Timeout: time.Second}, logger)
	request, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	response, err := client.Do(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	server.Close()
	if len(logger.error) != 1 || logger.error[0].metadata["status"] != http.StatusNotFound {
		t.Fatalf("error entries = %v", logger.error)
	}
	failing, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	if _, err := client.Do(context.Background(), failing); err == nil {
		t.Fatal("a closed server must fail")
	}
	if len(logger.error) != 2 || logger.error[1].err == nil || len(logger.debug) != 0 {
		t.Fatalf("debug=%d error=%v", len(logger.debug), logger.error)
	}
}

func TestMaskUrlHidesUserPasswordAndKeepsPlainUrls(t *testing.T) {
	target, _ := url.Parse("https://user:pass@example.com/a?q=1&token")
	if got := maskUrl(target); got != "https://user:xxxxx@example.com/a?q=1&token" {
		t.Fatalf("maskUrl = %q", got)
	}
	if maskUrl(nil) != "" {
		t.Fatal("nil url masks to empty")
	}
}

func TestClientFactoryBuildsLoggingClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	logger := &recordingLogger{}
	client := HttpClientFactory(logger).Build(Options{VerifySsl: true, Timeout: time.Second})
	request, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	response, err := client.Do(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	client.CloseIdleConnections()
	if response.StatusCode != http.StatusNoContent || len(logger.debug) != 1 {
		t.Fatalf("status=%d debug=%d", response.StatusCode, len(logger.debug))
	}
}
