package mcp

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type fakePublisher struct {
	statuses []entity.AIAccessStatus
	mu       sync.Mutex
}

func (f *fakePublisher) AccessStatusChanged(_ context.Context, status entity.AIAccessStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statuses = append(f.statuses, status)
	return nil
}
func (f *fakePublisher) ApprovalRequested(context.Context, entity.AIApproval) error { return nil }
func (f *fakePublisher) ApprovalResolved(context.Context, entity.AIApproval) error  { return nil }
func (f *fakePublisher) HistoryChanged(context.Context, string) error               { return nil }
func (f *fakePublisher) WorkspacesChanged(context.Context, string) error            { return nil }

type fakeAIAccess struct {
	records []dto.AIActivityRecord
	mu      sync.Mutex
}

func (f *fakeAIAccess) Status(context.Context) (*dto.AIAccessStatus, error) { return nil, nil }
func (f *fakeAIAccess) Start(context.Context) (*dto.AIAccessStatus, error)  { return nil, nil }
func (f *fakeAIAccess) Stop(context.Context) (*dto.AIAccessStatus, error)   { return nil, nil }
func (f *fakeAIAccess) Guard(_ context.Context, request dto.AIGuard) (*dto.AIGuarded, error) {
	guarded := dto.NewAIGuarded(strings.ReplaceAll(request.Text, "s3cret", "•••••"))
	return &guarded, nil
}
func (f *fakeAIAccess) Notify(context.Context, dto.AIChanges) error { return nil }
func (f *fakeAIAccess) RequestApproval(context.Context, dto.AIApprovalRequest) (*dto.AIApprovalOutcome, error) {
	return nil, nil
}
func (f *fakeAIAccess) AnswerApproval(context.Context, dto.AIApprovalAnswer) error { return nil }
func (f *fakeAIAccess) PendingApprovals(context.Context) ([]dto.AIApproval, error) { return nil, nil }
func (f *fakeAIAccess) DenyApprovals(context.Context, dto.AIApprovalDenial) error  { return nil }
func (f *fakeAIAccess) Record(_ context.Context, request dto.AIActivityRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records = append(f.records, request)
	return nil
}

type fakeWorkspaceList struct{ active string }

func (f fakeWorkspaceList) Handle(context.Context) (*dto.WorkspaceList, error) {
	return &dto.WorkspaceList{ActiveLocalId: f.active}, nil
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func newTestServer(t *testing.T, active string) (*Server, *fakePublisher, *fakeAIAccess, int) {
	t.Helper()
	publisher, access := &fakePublisher{}, &fakeAIAccess{}
	server := NewServer(publisher)
	server.Mount(Controllers{AIAccess: access, WorkspaceList: fakeWorkspaceList{active: active}})
	port := freePort(t)
	status := server.Apply(context.Background(), entity.AIAccessSettings{Enabled: true, Port: port})
	if !status.IsRunning() {
		t.Fatalf("Apply() = %+v", status)
	}
	t.Cleanup(func() { server.Stop(context.Background()) })
	return server, publisher, access, port
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

func TestServerListsTheToolsAndRunsThePipeline(t *testing.T) {
	server, publisher, access, port := newTestServer(t, "ws-1")
	closed := make(chan string, 1)
	server.OnSessionClosed(func(_ context.Context, sessionId string) { closed <- sessionId })
	session := connect(t, port, "Claude Code")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil || len(listed.Tools) != len(registry) || len(registry) != 24 {
		t.Fatalf("ListTools() = %d tools, %v", len(listed.Tools), err)
	}
	eventually(t, func() bool { return server.Status(context.Background()).Clients == 1 })
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "tree", Arguments: map[string]any{"action": "explode", "note": "s3cret"}})
	if err != nil || !result.IsError {
		t.Fatalf("CallTool() = %+v, %v", result, err)
	}
	text := result.Content[0].(*sdk.TextContent).Text
	if !strings.Contains(text, `"code":"MCP-07400"`) {
		t.Errorf("error text = %s", text)
	}
	access.mu.Lock()
	record := access.records[0]
	access.mu.Unlock()
	if record.Client != "Claude Code" || record.Action != "explode" || record.Outcome != entity.AIActivityError || record.Code != "MCP-07400" || record.LocalId != "ws-1" || !strings.Contains(record.Args, "s3cret") {
		t.Errorf("record = %+v", record)
	}
	_ = session.Close()
	select {
	case id := <-closed:
		if id == "" {
			t.Error("the closed session has no id")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the session close hook was not called")
	}
	eventually(t, func() bool { return server.Status(context.Background()).Clients == 0 })
	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	if len(publisher.statuses) < 3 || !publisher.statuses[0].IsRunning() {
		t.Errorf("published statuses = %+v", publisher.statuses)
	}
}

func TestServerNeedsAnActiveWorkspaceForScopedTools(t *testing.T) {
	_, _, access, port := newTestServer(t, "")
	session := connect(t, port, "")
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "tree", Arguments: map[string]any{"action": "list"}})
	if err != nil || !result.IsError || !strings.Contains(result.Content[0].(*sdk.TextContent).Text, "MCP-06404") {
		t.Fatalf("CallTool() = %+v, %v", result, err)
	}
	unscoped, _ := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "workspace", Arguments: map[string]any{"action": 3}})
	if !unscoped.IsError || !strings.Contains(unscoped.Content[0].(*sdk.TextContent).Text, "MCP-07400") {
		t.Errorf("a non-string action = %+v", unscoped)
	}
	access.mu.Lock()
	defer access.mu.Unlock()
	if len(access.records) != 2 || access.records[0].Client != entity.DefaultAIClientName {
		t.Errorf("records = %+v", access.records)
	}
	_ = session.Close()
}

func TestServerRejectsForeignHostsAndOrigins(t *testing.T) {
	_, _, _, port := newTestServer(t, "ws-1")
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

func TestServerReportsABusyPortAndStops(t *testing.T) {
	server, _, _, _ := newTestServer(t, "ws-1")
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	port := busy.Addr().(*net.TCPAddr).Port
	status := server.Apply(context.Background(), entity.AIAccessSettings{Enabled: true, Port: port})
	if status.State != "error" || status.Code != entity.PortInUseCode || status.Reason != "port "+strconv.Itoa(port)+" is in use" {
		t.Fatalf("busy port status = %+v", status)
	}
	if off := server.Apply(context.Background(), entity.AIAccessSettings{Enabled: false, Port: port}); off.State != "off" {
		t.Fatalf("disabled status = %+v", off)
	}
	if stopped := server.Stop(context.Background()); stopped.State != "off" {
		t.Fatalf("Stop() = %+v", stopped)
	}
}

func TestToolDefinitionsCarryTheContractHints(t *testing.T) {
	counts := map[effect]int{}
	for _, definition := range registry {
		counts[definition.effect]++
		annotations := definition.annotations()
		if annotations.OpenWorldHint == nil || *annotations.OpenWorldHint != definition.openWorld {
			t.Errorf("%s: open world hint %v", definition.name, annotations.OpenWorldHint)
		}
		if definition.effect != readOnly && (annotations.DestructiveHint == nil || *annotations.DestructiveHint != (definition.effect == destructive)) {
			t.Errorf("%s: destructive hint %v", definition.name, annotations.DestructiveHint)
		}
		schema := definition.inputSchema()
		if definition.hasActions() && schema["properties"] == nil {
			t.Errorf("%s: no action enum", definition.name)
		}
	}
	if counts[readOnly] != 12 || counts[write] != 2 || counts[destructive] != 10 {
		t.Errorf("hint counts = %v", counts)
	}
}
