---
name: Mock - API (HTTP Mock Server)
description: Use when creating or referencing the HTTP API mock server for BDD tests.
version: 1.0.0
---

# Mock: API (HTTP Mock Server)

Mocks external REST APIs using `httptest.NewServer`. Tracks all received requests (headers, query params, body) and returns configurable responses per method+path+index.

**Location:** `test/integration/mock/api.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync"
)

type ApiMock struct {
    mu              sync.RWMutex
    server          *httptest.Server
    responses       map[string]apiResponse         // key: "METHOD:path:index"
    defaultResponse map[string]apiResponse         // key: "METHOD:path"
    headers         map[string]map[string]string   // key: "METHOD:path:index" -> headers
    queries         map[string]map[string]string   // key: "METHOD:path:index" -> query params
    bodies          map[string][]byte              // key: "METHOD:path:index"
    callCount       map[string]int                 // key: "METHOD:path"
}

type apiResponse struct {
    statusCode  int
    body        []byte
    contentType string
}

func NewApiServer() *ApiMock {
    m := &ApiMock{}
    m.server = httptest.NewServer(http.HandlerFunc(m.handler))
    m.Start()
    return m
}

func (m *ApiMock) URL() string {
    return m.server.URL
}

func (m *ApiMock) Start() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.responses = make(map[string]apiResponse)
    m.defaultResponse = make(map[string]apiResponse)
    m.headers = make(map[string]map[string]string)
    m.queries = make(map[string]map[string]string)
    m.bodies = make(map[string][]byte)
    m.callCount = make(map[string]int)
}

func (m *ApiMock) SetResponse(index int, method, path string, status int, body []byte, contentType string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    key := fmt.Sprintf("%s:%s:%d", method, path, index)
    m.responses[key] = apiResponse{statusCode: status, body: body, contentType: contentType}
}

func (m *ApiMock) handler(w http.ResponseWriter, r *http.Request) {
    m.mu.Lock()
    method := r.Method
    path := r.URL.Path
    countKey := fmt.Sprintf("%s:%s", method, path)
    m.callCount[countKey]++
    index := m.callCount[countKey]

    // Store received request data
    reqKey := fmt.Sprintf("%s:%s:%d", method, path, index)
    hdrs := make(map[string]string)
    for k, v := range r.Header { hdrs[k] = v[0] }
    m.headers[reqKey] = hdrs

    qry := make(map[string]string)
    for k, v := range r.URL.Query() { qry[k] = v[0] }
    m.queries[reqKey] = qry

    body, _ := io.ReadAll(r.Body)
    m.bodies[reqKey] = body
    m.mu.Unlock()

    // Find matching response (exact key or wildcard path matching)
    resp := m.findResponse(method, path, index)
    if resp.contentType != "" {
        w.Header().Set("Content-Type", resp.contentType)
    } else {
        w.Header().Set("Content-Type", "application/json")
    }
    w.WriteHeader(resp.statusCode)
    w.Write(resp.body)
}

// Inspection methods for Then steps
func (m *ApiMock) GetRequestHeaders(index int, method, path string) map[string]string { ... }
func (m *ApiMock) GetRequestQueries(index int, method, path string) map[string]string { ... }
func (m *ApiMock) GetRequestBody(index int, method, path string) []byte { ... }
func (m *ApiMock) GetCallCount(method, path string) int { ... }
```

**Usage in step definitions:**
```go
// Given — set up mock response
test.api.SetResponse(1, "GET", "/api/resource/123", 200, responseBody, "application/json")

// Then — verify request was made correctly
headers := test.api.GetRequestHeaders(1, "POST", "/api/process")
body := test.api.GetRequestBody(1, "POST", "/api/process")
```

**Lifecycle:** The API mock is created once in `InitializeScenario`. The `Start()` method acts as its `Reset()` — it clears all responses, headers, queries, bodies, and call counts. Between BDD scenarios, `t.api.Start()` is called in `before()` to clear accumulated state. The mock instance is never recreated.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
