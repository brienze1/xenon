package mcpbridge

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

const (
	sessionIDHeader       = "Mcp-Session-Id"
	protocolVersionHeader = "Mcp-Protocol-Version"
	contentTypeJSON       = "application/json"
	contentTypeSSE        = "text/event-stream"
	connectTimeout        = 3 * time.Second
	closeSessionTimeout   = time.Second
	maxSSELineBytes       = 64 << 20
)

type statusError struct {
	status int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("Xenon answered HTTP %d", e.status)
}

type xenonClient struct {
	http            *http.Client
	endpoint        string
	sessionID       string
	protocolVersion string
	mu              sync.Mutex
}

func newXenonClient(endpoint string) *xenonClient {
	return &xenonClient{
		endpoint: endpoint,
		http: &http.Client{Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: connectTimeout}).DialContext,
		}},
	}
}

func (c *xenonClient) setProtocolVersion(version string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.protocolVersion = version
}

func (c *xenonClient) send(ctx context.Context, payload []byte, emit func(jsonrpc.Message)) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", contentTypeJSON)
	request.Header.Set("Accept", contentTypeJSON+", "+contentTypeSSE)
	c.setSessionHeaders(request.Header)

	response, err := c.http.Do(request)
	if err != nil {
		return errXenonClosed
	}
	defer func() { _ = response.Body.Close() }()
	c.rememberSession(response.Header.Get(sessionIDHeader))

	switch {
	case response.StatusCode == http.StatusNotFound && c.hasSession():
		return errXenonClosed
	case response.StatusCode < 200 || response.StatusCode >= 300:
		return emitErrorBody(response, emit)
	case response.StatusCode == http.StatusAccepted || response.StatusCode == http.StatusNoContent:
		return nil
	}

	mediaType, _, _ := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if mediaType == contentTypeSSE {
		return readEventStream(response.Body, emit)
	}
	return readJSONBody(response, emit)
}

func (c *xenonClient) closeSession() {
	if !c.hasSession() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), closeSessionTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.endpoint, nil)
	if err != nil {
		return
	}
	c.setSessionHeaders(request.Header)
	response, err := c.http.Do(request)
	if err != nil {
		return
	}
	_ = response.Body.Close()
}

func (c *xenonClient) setSessionHeaders(header http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessionID != "" {
		header.Set(sessionIDHeader, c.sessionID)
	}
	if c.protocolVersion != "" {
		header.Set(protocolVersionHeader, c.protocolVersion)
	}
}

func (c *xenonClient) rememberSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessionID == "" {
		c.sessionID = sessionID
	}
}

func (c *xenonClient) hasSession() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID != ""
}

func emitErrorBody(response *http.Response, emit func(jsonrpc.Message)) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errXenonClosed
	}
	message, err := jsonrpc.DecodeMessage(body)
	if _, isResponse := message.(*jsonrpc.Response); err == nil && isResponse {
		emit(message)
		return nil
	}
	return &statusError{status: response.StatusCode}
}

func readJSONBody(response *http.Response, emit func(jsonrpc.Message)) error {
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return errXenonClosed
	}
	message, err := jsonrpc.DecodeMessage(content)
	if err != nil {
		return &statusError{status: response.StatusCode}
	}
	emit(message)
	return nil
}

func readEventStream(body io.Reader, emit func(jsonrpc.Message)) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64<<10), maxSSELineBytes)
	var data []string
	dispatch := func() {
		if len(data) == 0 {
			return
		}
		if message, err := jsonrpc.DecodeMessage([]byte(strings.Join(data, "\n"))); err == nil {
			emit(message)
		}
		data = nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			dispatch()
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return errXenonClosed
	}
	dispatch()
	return nil
}
