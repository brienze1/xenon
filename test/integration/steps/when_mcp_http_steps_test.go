package steps

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerWhenMcpHttpSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I post to the MCP endpoint with the headers$`, t.iPostToTheMcpEndpointWithTheHeaders)
	ctx.Step(`^I post a body of (\d+) bytes to the MCP endpoint$`, t.iPostABodyOfBytesToTheMcpEndpoint)
}

func (t *testUtils) iPostToTheMcpEndpointWithTheHeaders(table *godog.Table) error {
	initialize := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"bdd-raw","version":"1.0.0"}}}`
	request, err := http.NewRequest(http.MethodPost, t.mcpEndpoint(), strings.NewReader(initialize))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	for _, row := range table.Rows[1:] {
		key, value := row.Cells[0].Value, t.substitute(row.Cells[1].Value)
		if strings.EqualFold(key, "Host") {
			request.Host = value
			continue
		}
		request.Header.Set(key, value)
	}
	return t.postToMcp(request)
}

func (t *testUtils) iPostABodyOfBytesToTheMcpEndpoint(size int) error {
	request, err := http.NewRequest(http.MethodPost, t.mcpEndpoint(), bytes.NewReader(generatedBody(size)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	return t.postToMcp(request)
}

func (t *testUtils) postToMcp(request *http.Request) error {
	started := time.Now()
	response, err := (&http.Client{Timeout: mcpCallTimeout}).Do(request)
	t.elapsed = time.Since(started)
	if err != nil {
		return fmt.Errorf("posting to the MCP endpoint %s: %w", t.mcpEndpoint(), err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	recorder := httptest.NewRecorder()
	for key, values := range response.Header {
		for _, value := range values {
			recorder.Header().Add(key, value)
		}
	}
	recorder.WriteHeader(response.StatusCode)
	_, _ = recorder.Write(body)
	t.routeResponse = recorder
	return nil
}
