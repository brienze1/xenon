package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

type Client interface {
	Do(ctx context.Context, request *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

type Options struct {
	Timeout         time.Duration
	FollowRedirects bool
	VerifySsl       bool
}

type ClientFactory func(options Options) Client

type customHttp struct {
	client *http.Client
}

// NewClient builds a client with its own transport, so every send opens fresh connections and
// reports full DNS, connect and TLS timings.
func NewClient(options Options) Client {
	transport := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: !options.VerifySsl},
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport, Timeout: options.Timeout}
	if !options.FollowRedirects {
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}
	return &customHttp{client: client}
}

func (c *customHttp) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	return c.client.Do(request.WithContext(ctx))
}

func (c *customHttp) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}
