package http

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		client := NewClient(Options{FollowRedirects: follow, VerifySsl: true})
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
		client := NewClient(Options{VerifySsl: verify, Timeout: time.Second})
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
