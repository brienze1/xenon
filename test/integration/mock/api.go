package mock

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"time"
)

type ApiMock struct {
	server    *httptest.Server
	responses map[string]apiResponse
	delays    map[string]time.Duration
	redirects map[string]apiRedirect
	requests  map[string]receivedRequest
	callCount map[string]int
	mu        sync.RWMutex
}

type apiResponse struct {
	contentType string
	body        []byte
	statusCode  int
}

type apiRedirect struct {
	target     string
	statusCode int
}

type receivedRequest struct {
	headers http.Header
	query   url.Values
	body    []byte
}

var apiInstance *ApiMock
var apiOnce sync.Once

func NewApiServer() *ApiMock {
	apiOnce.Do(func() {
		apiInstance = &ApiMock{}
		apiInstance.server = httptest.NewServer(http.HandlerFunc(apiInstance.handler))
		apiInstance.Start()
	})
	return apiInstance
}

func (m *ApiMock) URL() string {
	return m.server.URL
}

func (m *ApiMock) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = make(map[string]apiResponse)
	m.delays = make(map[string]time.Duration)
	m.redirects = make(map[string]apiRedirect)
	m.requests = make(map[string]receivedRequest)
	m.callCount = make(map[string]int)
}

func (m *ApiMock) SetResponse(index int, method, path string, status int, body []byte, contentType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[requestKey(method, path, index)] = apiResponse{statusCode: status, body: body, contentType: contentType}
}

func (m *ApiMock) SetDelay(method, path string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[routeKey(method, path)] = delay
}

func (m *ApiMock) SetRedirect(path, target string, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.redirects[path] = apiRedirect{target: target, statusCode: status}
}

func (m *ApiMock) GetCallCount(method, path string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount[routeKey(method, path)]
}

func (m *ApiMock) GetRequestHeader(index int, method, path, key string) (string, bool) {
	request, found := m.request(index, method, path)
	if !found || len(request.headers.Values(key)) == 0 {
		return "", false
	}
	return request.headers.Get(key), true
}

func (m *ApiMock) GetRequestQuery(index int, method, path, key string) (string, bool) {
	request, found := m.request(index, method, path)
	if !found || !request.query.Has(key) {
		return "", false
	}
	return request.query.Get(key), true
}

func (m *ApiMock) GetRequestBody(index int, method, path string) ([]byte, bool) {
	request, found := m.request(index, method, path)
	return request.body, found
}

func (m *ApiMock) request(index int, method, path string) (receivedRequest, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	request, found := m.requests[requestKey(method, path, index)]
	return request, found
}

func (m *ApiMock) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	m.mu.Lock()
	route := routeKey(r.Method, r.URL.Path)
	m.callCount[route]++
	index := m.callCount[route]
	m.requests[requestKey(r.Method, r.URL.Path, index)] = receivedRequest{headers: r.Header.Clone(), query: r.URL.Query(), body: body}
	delay := m.delays[route]
	redirect, redirected := m.redirects[r.URL.Path]
	response, configured := m.responses[requestKey(r.Method, r.URL.Path, index)]
	m.mu.Unlock()

	select {
	case <-time.After(delay):
	case <-r.Context().Done():
		return
	}
	if redirected {
		http.Redirect(w, r, redirect.target, redirect.statusCode)
		return
	}
	if !configured {
		response = apiResponse{statusCode: http.StatusOK, body: []byte("{}"), contentType: "application/json"}
	}
	w.Header().Set("Content-Type", response.contentType)
	w.WriteHeader(response.statusCode)
	_, _ = w.Write(response.body)
}

func routeKey(method, path string) string {
	return method + ":" + path
}

func requestKey(method, path string, index int) string {
	return fmt.Sprintf("%s:%s:%d", method, path, index)
}
