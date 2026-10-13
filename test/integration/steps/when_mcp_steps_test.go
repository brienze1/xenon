package steps

import (
	"bytes"
	"regexp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultMcpClientName = "bdd-client"
	mcpCallTimeout       = 30 * time.Second
	bridgeTimeout        = 20 * time.Second
	bridgePackage        = "internal/infra/mcpbridge"
)

var (
	bridgeBinaryOnce sync.Once
	bridgeBinary     string
	bridgeBuildError error
)

type bridgeRun struct {
	stdout   string
	stderr   string
	exitCode int
}

func (t *testUtils) registerWhenMcpSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the MCP client is named "([^"]*)"$`, t.theMcpClientIsNamed)
	ctx.Step(`^I call the MCP tool "([^"]*)" with$`, t.iCallTheMcpToolWith)
	ctx.Step(`^I call the MCP tool "([^"]*)"$`, t.iCallTheMcpTool)
	ctx.Step(`^I call the MCP tool "([^"]*)" (\d+) times with$`, t.iCallTheMcpToolTimesWith)
	ctx.Step(`^I start calling the MCP tool "([^"]*)" in the background with$`, t.iStartCallingTheMcpToolInTheBackgroundWith)
	ctx.Step(`^I start calling the MCP tool "([^"]*)" in the background as "([^"]*)" with$`, t.iStartCallingTheMcpToolInTheBackgroundAsWith)
	ctx.Then(`^the background call "([^"]*)" should finish within (\d+) ms$`, t.theNamedBackgroundCallShouldFinishWithin)
	ctx.Step(`^I list the MCP tools$`, t.iListTheMcpTools)
	ctx.Step(`^the MCP session is closed$`, t.theMcpSessionIsClosed)
	ctx.Step(`^I post to the MCP endpoint with the headers$`, t.iPostToTheMcpEndpointWithTheHeaders)
	ctx.Step(`^I post a body of (\d+) bytes to the MCP endpoint$`, t.iPostABodyOfBytesToTheMcpEndpoint)
	ctx.Step(`^I run the stdio bridge with the input$`, t.iRunTheStdioBridgeWithTheInput)
	ctx.Then(`^the MCP endpoint should refuse connections$`, t.theMcpEndpointShouldRefuseConnections)
	ctx.Then(`^the bridge exit code should be (\d+)$`, t.theBridgeExitCodeShouldBe)
	ctx.Then(`^the bridge output should contain "([^"]*)"$`, t.theBridgeOutputShouldContain)
	ctx.Then(`^the bridge output should not contain "([^"]*)"$`, t.theBridgeOutputShouldNotContain)
	ctx.Then(`^the bridge output should match "([^"]*)"$`, t.theBridgeOutputShouldMatch)
	ctx.Then(`^the bridge error output should contain "([^"]*)"$`, t.theBridgeErrorOutputShouldContain)
}

func (t *testUtils) prepareMcp() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	if err := listener.Close(); err != nil {
		return err
	}
	t.mcpClientName = defaultMcpClientName
	t.mcpSessions = map[string]*mcp.ClientSession{}
	t.bridge = nil
	t.vars["mcp_port"] = port
	t.vars["mcp_url"] = "http://127.0.0.1:" + port + "/mcp"
	if _, saved := t.environment["XENON_MCP_PORT"]; !saved {
		t.environment["XENON_MCP_PORT"] = os.Getenv("XENON_MCP_PORT")
	}
	return os.Setenv("XENON_MCP_PORT", port)
}

func (t *testUtils) closeMcp() {
	for name, session := range t.mcpSessions {
		_ = session.Close()
		delete(t.mcpSessions, name)
	}
}

func (t *testUtils) mcpEndpoint() string {
	return "http://127.0.0.1:" + t.vars["mcp_port"] + "/mcp"
}

func (t *testUtils) session(ctx context.Context) (*mcp.ClientSession, error) {
	if session, open := t.mcpSessions[t.mcpClientName]; open {
		return session, nil
	}
	client := mcp.NewClient(&mcp.Implementation{Name: t.mcpClientName, Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: t.mcpEndpoint(), MaxRetries: -1}, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting an MCP client to %s: %w", t.mcpEndpoint(), err)
	}
	t.mcpSessions[t.mcpClientName] = session
	return session, nil
}

func (t *testUtils) theMcpClientIsNamed(name string) error {
	t.mcpClientName = t.substitute(name)
	return nil
}

func (t *testUtils) theMcpSessionIsClosed() error {
	if session, open := t.mcpSessions[t.mcpClientName]; open {
		delete(t.mcpSessions, t.mcpClientName)
		return session.Close()
	}
	return nil
}

func (t *testUtils) iCallTheMcpToolWith(tool string, arguments *godog.DocString) error {
	return t.callMcpTool(tool, arguments.Content)
}

func (t *testUtils) iCallTheMcpTool(tool string) error {
	return t.callMcpTool(tool, "")
}

func (t *testUtils) iCallTheMcpToolTimesWith(tool string, times int, arguments *godog.DocString) error {
	for call := 0; call < times; call++ {
		if err := t.callMcpTool(tool, arguments.Content); err != nil {
			return err
		}
		if t.callError != nil {
			return fmt.Errorf("call %d of the MCP tool %s failed: %s", call+1, tool, t.callError)
		}
	}
	return nil
}

func (t *testUtils) callMcpTool(tool, arguments string) error {
	t.response, t.callError = nil, nil
	invoke, err := t.prepareMcpCall(tool, arguments)
	if err != nil {
		return err
	}
	result, err := invoke()
	t.response, t.callError, t.elapsed = result.response, result.callError, result.elapsed
	return err
}

func (t *testUtils) iStartCallingTheMcpToolInTheBackgroundWith(tool string, arguments *godog.DocString) error {
	invoke, err := t.prepareMcpCall(tool, arguments.Content)
	if err != nil {
		return err
	}
	call := &backgroundCall{done: make(chan struct{}), started: time.Now()}
	t.background = call
	go func() {
		call.result, call.err = invoke()
		close(call.done)
	}()
	return nil
}

func (t *testUtils) iStartCallingTheMcpToolInTheBackgroundAsWith(tool, name string, arguments *godog.DocString) error {
	if err := t.iStartCallingTheMcpToolInTheBackgroundWith(tool, arguments); err != nil {
		return err
	}
	t.backgrounds[name] = t.background
	return nil
}

func (t *testUtils) theNamedBackgroundCallShouldFinishWithin(name string, milliseconds int) error {
	call, found := t.backgrounds[name]
	if !found {
		return fmt.Errorf("no background call named %q was started", name)
	}
	t.background = call
	return t.theBackgroundCallShouldFinishWithin(milliseconds)
}

func (t *testUtils) prepareMcpCall(tool, arguments string) (func() (callResult, error), error) {
	params := &mcp.CallToolParams{Name: tool, Arguments: map[string]any{}}
	if strings.TrimSpace(arguments) != "" {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(t.substituteJson(arguments)), &decoded); err != nil {
			return nil, fmt.Errorf("invalid arguments for the MCP tool %s: %w", tool, err)
		}
		params.Arguments = decoded
	}
	connectCtx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
	defer cancel()
	session, err := t.session(connectCtx)
	if err != nil {
		return nil, err
	}
	return func() (callResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
		defer cancel()
		started := time.Now()
		result, err := session.CallTool(ctx, params)
		elapsed := time.Since(started)
		if err != nil {
			return callResult{elapsed: elapsed}, fmt.Errorf("calling the MCP tool %s: %w", tool, err)
		}
		text := toolText(result)
		if result.IsError {
			return callResult{callError: wrapToolError(text), elapsed: elapsed}, nil
		}
		return callResult{response: []byte(text), elapsed: elapsed}, nil
	}, nil
}

func toolText(result *mcp.CallToolResult) string {
	var builder strings.Builder
	for _, content := range result.Content {
		if text, isText := content.(*mcp.TextContent); isText {
			builder.WriteString(text.Text)
		}
	}
	return builder.String()
}

func wrapToolError(text string) []byte {
	var payload any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		payload = map[string]any{"message": text}
	}
	wrapped, _ := json.Marshal(map[string]any{"error": payload})
	return wrapped
}

func (t *testUtils) iListTheMcpTools() error {
	t.response, t.callError = nil, nil
	ctx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
	defer cancel()
	started := time.Now()
	session, err := t.session(ctx)
	if err != nil {
		return err
	}
	listed, err := session.ListTools(ctx, nil)
	t.elapsed = time.Since(started)
	if err != nil {
		return fmt.Errorf("listing the MCP tools: %w", err)
	}
	t.response, err = json.Marshal(listed.Tools)
	return err
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

func (t *testUtils) theMcpEndpointShouldRefuseConnections() error {
	connection, err := net.DialTimeout("tcp", "127.0.0.1:"+t.vars["mcp_port"], time.Second)
	if err == nil {
		_ = connection.Close()
		return fmt.Errorf("the MCP endpoint %s accepted a connection", t.mcpEndpoint())
	}
	return nil
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

func (t *testUtils) iRunTheStdioBridgeWithTheInput(input *godog.DocString) error {
	binary, err := buildBridge()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), bridgeTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "mcp")
	command.Env = append(os.Environ(), "XENON_DATA_DIR="+t.dataDir)
	command.Stdin = strings.NewReader(t.substitute(input.Content) + "\n")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runErr := command.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("the stdio bridge did not exit within %s; stderr: %s", bridgeTimeout, stderr.String())
	}
	run := &bridgeRun{stdout: stdout.String(), stderr: stderr.String()}
	var exitErr *exec.ExitError
	switch {
	case errors.As(runErr, &exitErr):
		run.exitCode = exitErr.ExitCode()
	case runErr != nil:
		return runErr
	}
	t.bridge = run
	return nil
}

func buildBridge() (string, error) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(bridgePackage))); err != nil {
		return "", fmt.Errorf("the stdio bridge is not implemented yet (%s is missing)", bridgePackage)
	}
	bridgeBinaryOnce.Do(func() {
		folder, err := os.MkdirTemp("", "xenon-bridge-")
		if err != nil {
			bridgeBuildError = err
			return
		}
		bridgeBinary = filepath.Join(folder, "xenon")
		build := exec.Command("go", "build", "-o", bridgeBinary, "./cmd/xenon")
		build.Dir = root
		if output, err := build.CombinedOutput(); err != nil {
			bridgeBuildError = fmt.Errorf("building ./cmd/xenon: %w: %s", err, output)
		}
	})
	return bridgeBinary, bridgeBuildError
}

func (t *testUtils) bridgeResult() (*bridgeRun, error) {
	if t.bridge == nil {
		return nil, fmt.Errorf("the stdio bridge has not been run")
	}
	return t.bridge, nil
}

func (t *testUtils) theBridgeExitCodeShouldBe(code int) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if run.exitCode != code {
		return fmt.Errorf("expected exit code %d, got %d; stderr: %s", code, run.exitCode, run.stderr)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if !strings.Contains(run.stdout, t.substitute(text)) {
		return fmt.Errorf("the bridge output does not contain %q: %s", t.substitute(text), run.stdout)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldNotContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if strings.Contains(run.stdout, t.substitute(text)) {
		return fmt.Errorf("the bridge output contains %q: %s", t.substitute(text), run.stdout)
	}
	return nil
}

func (t *testUtils) theBridgeErrorOutputShouldContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if !strings.Contains(run.stderr, t.substitute(text)) {
		return fmt.Errorf("the bridge error output does not contain %q: %s", t.substitute(text), run.stderr)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldMatch(pattern string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	expression, err := regexp.Compile(t.substitute(pattern))
	if err != nil {
		return err
	}
	if !expression.MatchString(run.stdout) {
		return fmt.Errorf("the bridge output does not match %q: %s", pattern, run.stdout)
	}
	return nil
}
