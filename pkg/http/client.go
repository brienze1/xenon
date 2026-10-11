package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	Do(ctx context.Context, request *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

type Options struct {
	Jar             http.CookieJar
	Timeout         time.Duration
	FollowRedirects bool
	VerifySsl       bool
}

type ClientFactory interface {
	Build(options Options) Client
}

type loggerAdapter interface {
	Debug(ctx context.Context, message string, metadata ...any)
	Error(ctx context.Context, err error, message string, metadata ...any)
}

type customHttp struct {
	client *http.Client
	logger loggerAdapter
}

type clientFactory struct {
	logger loggerAdapter
}

func HttpClientFactory(logger loggerAdapter) ClientFactory {
	return &clientFactory{logger: logger}
}

func (f *clientFactory) Build(options Options) Client {
	return HttpClient(options, f.logger)
}

// HttpClient creates a client with logging and a dedicated transport.
func HttpClient(options Options, logger loggerAdapter) Client {
	transport := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: !options.VerifySsl},
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport, Timeout: options.Timeout, Jar: options.Jar}
	if !options.FollowRedirects {
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}
	return &customHttp{client: client, logger: logger}
}

func (c *customHttp) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	started := time.Now()
	response, err := c.client.Do(request.WithContext(ctx))
	metadata := requestMetadata(request, time.Since(started))
	if err != nil {
		c.logger.Error(ctx, err, "HTTP request failed", metadata)
		return response, err
	}
	metadata["status"] = response.StatusCode
	metadata["response_headers"] = maskHeaders(response.Header)
	if response.StatusCode >= http.StatusBadRequest {
		c.logger.Error(ctx, fmt.Errorf("http status %d", response.StatusCode), "HTTP request returned an error status", metadata)
		return response, nil
	}
	c.logger.Debug(ctx, "HTTP request finished", metadata)
	return response, nil
}

func (c *customHttp) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

func requestMetadata(request *http.Request, duration time.Duration) map[string]any {
	return map[string]any{
		"method":      request.Method,
		"url":         maskUrl(request.URL),
		"headers":     maskHeaders(request.Header),
		"duration_ms": duration.Milliseconds(),
	}
}
