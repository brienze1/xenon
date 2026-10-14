package mcpbridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const (
	testSessionID      = "session-1"
	initializeLine     = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test-client","version":"1"}}}`
	initializedLine    = `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	toolsListLine      = `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	notRunningMessage  = "Xenon is not running — start Xenon with AI access on, then retry.\n"
	accessOffMessage   = "Xenon AI access is off — turn it on in Xenon → Settings → AI access (MCP).\n"
	closedMessage      = "Xenon closed the connection.\n"
	foreignPortMessage = "Port %d is in use by another program — change the port in Xenon → Settings → AI access (MCP).\n"
)

type fakeServer struct {
	mu            sync.Mutex
	name          string
	abortTools    bool
	toolsSession  string
	toolsProtocol string
}

func (f *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var message struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch message.Method {
	case "initialize":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(sessionIDHeader, testSessionID)
		_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-06-18","capabilities":{"tools":{}},"serverInfo":{"name":%q,"version":"1"}}}`, message.ID, f.name)
	case "tools/list":
		if f.abortTools {
			panic(http.ErrAbortHandler)
		}
		f.mu.Lock()
		f.toolsSession = r.Header.Get(sessionIDHeader)
		f.toolsProtocol = r.Header.Get(protocolVersionHeader)
		f.mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"tools\":[{\"name\":\"workspace_read\",\"inputSchema\":{\"type\":\"object\"}},{\"name\":\"send\",\"inputSchema\":{\"type\":\"object\"}}]}}\n\n", message.ID)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}

func startFakeServer(t *testing.T, server *fakeServer) int {
	t.Helper()
	httpServer := httptest.NewServer(server)
	t.Cleanup(httpServer.Close)
	return portOf(t, httpServer.URL)
}

func portOf(t *testing.T, rawURL string) int {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	var port int
	if _, err := fmt.Sscan(parsed.Port(), &port); err != nil {
		t.Fatal(err)
	}
	return port
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func useSettings(t *testing.T, appJSON string, overridePort int) {
	t.Helper()
	dataDir := t.TempDir()
	if appJSON != "" {
		if err := os.WriteFile(filepath.Join(dataDir, "app.json"), []byte(appJSON), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("XENON_DATA_DIR", dataDir)
	t.Setenv("XENON_MCP_PORT", fmt.Sprint(overridePort))
}

func runBridge(lines ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(strings.Join(lines, "\n") + "\n")
	code := Run(stdin, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunForwardsSessionToXenon(t *testing.T) {
	server := &fakeServer{name: xenonServerName}
	useSettings(t, "", startFakeServer(t, server))

	code, stdout, stderr := runBridge(initializeLine, initializedLine, toolsListLine)

	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 protocol lines, got %q", stdout)
	}
	if !strings.Contains(lines[0], `"name":"xenon"`) || !strings.Contains(lines[1], "workspace_read") || !strings.Contains(lines[1], `"send"`) {
		t.Fatalf("unexpected output %q", stdout)
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.toolsSession != testSessionID || server.toolsProtocol != "2025-06-18" {
		t.Fatalf("session headers not carried: %q %q", server.toolsSession, server.toolsProtocol)
	}
}

func TestRunRefusesForeignServer(t *testing.T) {
	port := startFakeServer(t, &fakeServer{name: "not-xenon"})
	useSettings(t, "", port)

	code, stdout, stderr := runBridge(initializeLine, toolsListLine)

	if code != 1 || stdout != "" || stderr != fmt.Sprintf(foreignPortMessage, port) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunRefusesNonMCPServer(t *testing.T) {
	httpServer := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(httpServer.Close)
	port := portOf(t, httpServer.URL)
	useSettings(t, "", port)

	code, stdout, stderr := runBridge(initializeLine)

	if code != 1 || stdout != "" || stderr != fmt.Sprintf(foreignPortMessage, port) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunStopsWhenAccessIsOff(t *testing.T) {
	useSettings(t, `{"settings":{"aiAccess":{"enabled":false}}}`, startFakeServer(t, &fakeServer{name: xenonServerName}))

	code, stdout, stderr := runBridge(initializeLine)

	if code != 1 || stdout != "" || stderr != accessOffMessage {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunReportsXenonNotRunning(t *testing.T) {
	useSettings(t, "", freePort(t))

	code, stdout, stderr := runBridge(initializeLine)

	if code != 1 || stdout != "" || stderr != notRunningMessage {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunPrefersSavedPortOverOverride(t *testing.T) {
	port := startFakeServer(t, &fakeServer{name: xenonServerName})
	useSettings(t, fmt.Sprintf(`{"settings":{"aiAccess":{"enabled":true,"port":%d}}}`, port), freePort(t))

	code, stdout, stderr := runBridge(initializeLine)

	if code != 0 || stderr != "" || !strings.Contains(stdout, `"name":"xenon"`) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunReportsXenonClosingTheConnection(t *testing.T) {
	useSettings(t, "", startFakeServer(t, &fakeServer{name: xenonServerName, abortTools: true}))

	code, _, stderr := runBridge(initializeLine, toolsListLine)

	if code != 1 || stderr != closedMessage {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

func TestRunReadsUnreadableSettingsAsDefaults(t *testing.T) {
	useSettings(t, `not json`, startFakeServer(t, &fakeServer{name: xenonServerName}))

	code, stdout, stderr := runBridge(initializeLine)

	if code != 0 || stderr != "" || !strings.Contains(stdout, `"name":"xenon"`) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
