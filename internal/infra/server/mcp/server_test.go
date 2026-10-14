package mcp

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/middleware"
)

type receivedCall struct {
	call dto.MCPCall
	tool string
}

type fakeCallController struct {
	calls  []receivedCall
	closed chan string
	mu     sync.Mutex
}

func newFakeCallController() *fakeCallController {
	return &fakeCallController{closed: make(chan string, 4)}
}

func (f *fakeCallController) Call(_ context.Context, call dto.MCPCall, tool adapter.MCPToolController) dto.MCPOutcome {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, receivedCall{call: call, tool: tool.Tool().Name})
	return dto.NewMCPOutcome(`{"code":"MCP-07400"}`, true)
}

func (f *fakeCallController) CloseSession(_ context.Context, session dto.MCPSession) error {
	f.closed <- session.Id
	return nil
}

type fakeToolController struct {
	tool entity.AITool
}

func (f fakeToolController) Tool() dto.MCPTool {
	return dto.NewMCPTool(f.tool)
}

func (f fakeToolController) Handle(context.Context, dto.MCPToolCall) (dto.MCPToolResult, error) {
	return dto.MCPToolResult{}, nil
}

func fakeTool(name string) adapter.MCPToolController {
	return fakeToolController{tool: entity.AITool{Name: name, Description: name, Effect: aitooleffect.ReadOnly}}
}

func fakeControllers(call adapter.MCPCallController) Controllers {
	return Controllers{
		Call:          call,
		Workspace:     fakeTool("workspace"),
		WorkspaceEdit: fakeTool("workspace_edit"),
		Tree:          fakeTool("tree"),
		TreeEdit:      fakeTool("tree_edit"),
		Env:           fakeTool("env"),
		EnvEdit:       fakeTool("env_edit"),
		History:       fakeTool("history"),
		Docs:          fakeTool("docs"),
		DocsEdit:      fakeTool("docs_edit"),
		Response:      fakeTool("response"),
		Snippet:       fakeTool("snippet"),
		Send:          fakeTool("send"),
		Cookies:       fakeTool("cookies"),
		CookiesEdit:   fakeTool("cookies_edit"),
		Console:       fakeTool("console"),
		Git:           fakeTool("git"),
		GitEdit:       fakeTool("git_edit"),
		GitPush:       fakeTool("git_push"),
		Scripts:       fakeTool("scripts"),
		Settings:      fakeTool("settings"),
		SettingsEdit:  fakeTool("settings_edit"),
		Import:        fakeTool("import"),
		Export:        fakeTool("export"),
		Delete:        fakeTool("delete"),
	}
}

func realToolControllers() []adapter.MCPToolController {
	return []adapter.MCPToolController{
		controller.NewMCPWorkspace(nil, nil, nil),
		controller.NewMCPWorkspaceEdit(nil, nil, nil, nil, nil, nil),
		controller.NewMCPTree(nil, nil, nil, nil, nil, nil, nil),
		controller.NewMCPTreeEdit(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
		controller.NewMCPEnv(nil, nil, nil, nil, nil, nil),
		controller.NewMCPEnvEdit(nil, nil, nil, nil, nil),
		controller.NewMCPHistory(nil, nil),
		controller.NewMCPDocs(nil, nil, nil),
		controller.NewMCPDocsEdit(nil, nil, nil, nil),
		controller.NewMCPResponse(nil, nil, nil),
		controller.NewMCPSnippet(nil, nil, nil, nil),
		controller.NewMCPSend(nil, nil, nil, nil, nil, nil, nil, nil),
		controller.NewMCPCookies(nil, nil),
		controller.NewMCPCookiesEdit(nil, nil),
		controller.NewMCPConsole(nil, nil),
		controller.NewMCPGit(nil, nil, nil),
		controller.NewMCPGitEdit(nil, nil, nil, nil, nil, nil, nil, nil),
		controller.NewMCPGitPush(nil, nil),
		controller.NewMCPScripts(nil, nil),
		controller.NewMCPSettings(nil, nil),
		controller.NewMCPSettingsEdit(nil, nil),
		controller.NewMCPImport(nil, nil),
		controller.NewMCPExport(nil, nil, nil, nil, nil),
		controller.NewMCPDelete(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	return busy.Addr().(*net.TCPAddr).Port
}

func newTestListener(t *testing.T, call adapter.MCPCallController) serverAdapter.MCPListener {
	t.Helper()
	testListener := NewListener()
	testListener.Route(NewRoutes(fakeControllers(call), middleware.NewLoopbackGuardMiddleware(), middleware.NewMCPSessionClosingMiddleware(call)))
	t.Cleanup(testListener.Close)
	return testListener
}

func listen(t *testing.T, testListener serverAdapter.MCPListener, clientsChanged func(ctx context.Context)) int {
	t.Helper()
	port := freePort(t)
	testListener.Watch(clientsChanged)
	if err := testListener.Listen(port); err != nil {
		t.Fatalf("Listen() = %v", err)
	}
	return port
}

func connect(t *testing.T, port int, name string) *sdk.ClientSession {
	t.Helper()
	client := sdk.NewClient(&sdk.Implementation{Name: name, Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &sdk.StreamableClientTransport{Endpoint: entity.AIAccessURL(port), MaxRetries: -1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestListenerServesTheToolsThroughTheCallController(t *testing.T) {
	call := newFakeCallController()
	testListener := newTestListener(t, call)
	var changes atomic.Int32
	port := listen(t, testListener, func(context.Context) { changes.Add(1) })
	session := connect(t, port, "Claude Code")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil || len(listed.Tools) != 24 {
		t.Fatalf("ListTools() = %d tools, %v", len(listed.Tools), err)
	}
	eventually(t, func() bool { return testListener.Clients() == 1 })
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "tree", Arguments: map[string]any{"action": "explode", "note": "s3cret"}})
	if err != nil || !result.IsError || result.Content[0].(*sdk.TextContent).Text != `{"code":"MCP-07400"}` {
		t.Fatalf("CallTool() = %+v, %v", result, err)
	}
	call.mu.Lock()
	received := call.calls[0]
	call.mu.Unlock()
	if received.tool != "tree" || received.call.Client != "Claude Code" || received.call.Session == "" || !strings.Contains(string(received.call.Arguments), "s3cret") {
		t.Errorf("received call = %+v", received)
	}
	_ = session.Close()
	select {
	case id := <-call.closed:
		if id != received.call.Session {
			t.Errorf("closed session %q, want %q", id, received.call.Session)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the call controller was not told the session closed")
	}
	eventually(t, func() bool { return testListener.Clients() == 0 })
	if changes.Load() < 2 {
		t.Errorf("clients changed %d times, want at least 2", changes.Load())
	}
}

func TestListenerRejectsForeignHostsAndOrigins(t *testing.T) {
	port := listen(t, newTestListener(t, newFakeCallController()), nil)
	cases := []struct {
		header string
		value  string
		status int
	}{
		{header: "Host", value: "evil.example:" + strconv.Itoa(port), status: http.StatusForbidden},
		{header: "Origin", value: "null", status: http.StatusForbidden},
		{header: "Origin", value: "http://localhost:" + strconv.Itoa(port), status: http.StatusOK},
	}
	for _, test := range cases {
		request, _ := http.NewRequest(http.MethodPost, entity.AIAccessURL(port), strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"raw","version":"1"}}}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json, text/event-stream")
		if test.header == "Host" {
			request.Host = test.value
		} else {
			request.Header.Set(test.header, test.value)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != test.status {
			t.Errorf("%s %s: status %d, want %d", test.header, test.value, response.StatusCode, test.status)
		}
	}
}

func TestListenerReportsABusyPortAndCloses(t *testing.T) {
	testListener := newTestListener(t, newFakeCallController())
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	if err := testListener.Listen(busy.Addr().(*net.TCPAddr).Port); err == nil {
		t.Fatal("Listen() on a busy port must fail")
	}
	testListener.Close()
	if clients := testListener.Clients(); clients != 0 {
		t.Fatalf("Clients() = %d", clients)
	}
}

func TestToolDefinitionsCarryTheContractHints(t *testing.T) {
	counts := map[enums.AIToolEffect]int{}
	for _, tool := range realToolControllers() {
		definition := tool.Tool()
		spec := definition.ToEntity()
		counts[spec.Effect]++
		sdkTool := newToolDefinition(definition)
		annotations := sdkTool.Annotations
		if annotations.OpenWorldHint == nil || *annotations.OpenWorldHint != spec.OpenWorld {
			t.Errorf("%s: open world hint %v", spec.Name, annotations.OpenWorldHint)
		}
		if spec.Effect != aitooleffect.ReadOnly && (annotations.DestructiveHint == nil || *annotations.DestructiveHint != (spec.Effect == aitooleffect.Destructive)) {
			t.Errorf("%s: destructive hint %v", spec.Name, annotations.DestructiveHint)
		}
		schema := sdkTool.InputSchema.(map[string]any)
		if len(spec.Actions) > 0 && schema["properties"] == nil {
			t.Errorf("%s: no action enum", spec.Name)
		}
	}
	if counts[aitooleffect.ReadOnly] != 12 || counts[aitooleffect.Write] != 2 || counts[aitooleffect.Destructive] != 10 {
		t.Errorf("hint counts = %v", counts)
	}
}
