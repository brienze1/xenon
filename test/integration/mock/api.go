package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ApiMock struct {
	server          *httptest.Server
	tlsServer       *httptest.Server
	responses       map[string]apiResponse
	responseHeaders map[string]http.Header
	delays          map[string]time.Duration
	redirects       map[string]apiRedirect
	echoes          map[string]int
	requests        map[string]receivedRequest
	callCount       map[string]int
	sequence        []string
	mu              sync.RWMutex
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
	at       time.Time
	headers  http.Header
	query    url.Values
	rawQuery string
	body     []byte
}

var apiInstance *ApiMock
var apiOnce sync.Once

func NewApiServer() *ApiMock {
	apiOnce.Do(func() {
		apiInstance = &ApiMock{}
		apiInstance.server = httptest.NewUnstartedServer(http.HandlerFunc(apiInstance.handler))
		apiInstance.tlsServer = httptest.NewUnstartedServer(http.HandlerFunc(apiInstance.handler))
		apiInstance.Reset()
		apiInstance.Start()
	})
	return apiInstance
}

func (m *ApiMock) URL() string {
	return m.server.URL
}

func (m *ApiMock) TLSURL() string {
	return m.tlsServer.URL
}

func (m *ApiMock) Start() {
	m.server.Start()
	m.tlsServer.StartTLS()
}

func (m *ApiMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = make(map[string]apiResponse)
	m.responseHeaders = make(map[string]http.Header)
	m.delays = make(map[string]time.Duration)
	m.redirects = make(map[string]apiRedirect)
	m.echoes = make(map[string]int)
	m.requests = make(map[string]receivedRequest)
	m.callCount = make(map[string]int)
	m.sequence = nil
}

func (m *ApiMock) SetResponse(index int, method, path string, status int, body []byte, contentType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[requestKey(method, path, index)] = apiResponse{statusCode: status, body: body, contentType: contentType}
}

func (m *ApiMock) AddResponseHeader(path, key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.responseHeaders[path] == nil {
		m.responseHeaders[path] = http.Header{}
	}
	m.responseHeaders[path].Add(key, value)
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

func (m *ApiMock) SetEcho(path string, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.echoes[path] = status
}

func writeEcho(w http.ResponseWriter, r *http.Request, body []byte, status int) {
	headers := map[string]string{}
	for key, values := range r.Header {
		joined := strings.Join(values, ", ")
		headers[key] = joined
		w.Header().Set("X-Echo-"+key, joined)
	}
	echo, _ := json.Marshal(map[string]any{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"headers": headers,
		"body":    string(body),
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(echo)
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

func (m *ApiMock) GetRequestRawQuery(index int, method, path string) (string, bool) {
	request, found := m.request(index, method, path)
	return request.rawQuery, found
}

func (m *ApiMock) GetRequestFormValue(index int, method, path, key string) (string, bool) {
	form, found := m.form(index, method, path)
	if !found || len(form.Value[key]) == 0 {
		return "", false
	}
	return form.Value[key][0], true
}

func (m *ApiMock) GetRequestFormFile(index int, method, path, key string) (string, []byte, bool) {
	form, found := m.form(index, method, path)
	if !found || len(form.File[key]) == 0 {
		return "", nil, false
	}
	header := form.File[key][0]
	file, err := header.Open()
	if err != nil {
		return "", nil, false
	}
	defer func() { _ = file.Close() }()
	content, _ := io.ReadAll(file)
	return header.Filename, content, true
}

func (m *ApiMock) form(index int, method, path string) (*multipart.Form, bool) {
	received, found := m.request(index, method, path)
	if !found {
		return nil, false
	}
	request, _ := http.NewRequest(method, "/", bytes.NewReader(received.body))
	request.Header = received.headers.Clone()
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		return nil, false
	}
	return request.MultipartForm, true
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
	m.requests[requestKey(r.Method, r.URL.Path, index)] = receivedRequest{headers: r.Header.Clone(), query: r.URL.Query(), rawQuery: r.URL.RawQuery, body: body, at: time.Now()}
	m.sequence = append(m.sequence, r.Method+" "+r.URL.Path)
	delay := m.delays[route]
	redirect, redirected := m.redirects[r.URL.Path]
	echoStatus, echoed := m.echoes[r.URL.Path]
	response, configured := m.responses[requestKey(r.Method, r.URL.Path, index)]
	extraHeaders := m.responseHeaders[r.URL.Path].Clone()
	m.mu.Unlock()

	select {
	case <-time.After(delay):
	case <-r.Context().Done():
		return
	}
	for key, values := range extraHeaders {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if redirected {
		http.Redirect(w, r, redirect.target, redirect.statusCode)
		return
	}
	if echoed {
		writeEcho(w, r, body, echoStatus)
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

func (m *ApiMock) GetRequest(index int, method, path string) (http.Header, url.Values, []byte, bool) {
	request, found := m.request(index, method, path)
	return request.headers, request.query, request.body, found
}

func (m *ApiMock) Sequence() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.sequence...)
}

func (m *ApiMock) GetReceivedAt(index int, method, path string) (time.Time, bool) {
	request, found := m.request(index, method, path)
	return request.at, found
}
