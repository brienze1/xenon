package controller

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type recordedAIActivity struct {
	localId  string
	client   string
	session  string
	activity entity.AIActivity
}

type fakeAIActivityRecorder struct {
	records []recordedAIActivity
	mu      sync.Mutex
}

func (f *fakeAIActivityRecorder) Record(_ context.Context, localId, client, session string, activity entity.AIActivity) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records = append(f.records, recordedAIActivity{localId: localId, client: client, session: session, activity: activity})
	return nil
}

type maskingAIOutputGuard struct{}

func (maskingAIOutputGuard) Guard(_ context.Context, _ string, text string) (string, error) {
	return strings.ReplaceAll(text, "s3cret", "•••••"), nil
}

type silentAIChangeNotifier struct{}

func (silentAIChangeNotifier) Notify(context.Context, string, []entity.AIChange) error { return nil }

type deniedSessions struct {
	ids []string
}

func (d *deniedSessions) Request(_ context.Context, approval entity.AIApproval, _ string) (entity.AIApproval, error) {
	return approval, nil
}
func (d *deniedSessions) Answer(context.Context, string, bool) error { return nil }
func (d *deniedSessions) Deny(_ context.Context, sessionId string) error {
	d.ids = append(d.ids, sessionId)
	return nil
}

type activeWorkspace struct{ localId string }

func (a activeWorkspace) List(context.Context) (*entity.AppRegistry, error) {
	return &entity.AppRegistry{ActiveLocalID: a.localId}, nil
}

type fakeAISessions struct {
	localId string
	opened  []string
	closed  []string
}

func (f *fakeAISessions) Open(_ context.Context, sessionId string) (*entity.AISessionContext, error) {
	f.opened = append(f.opened, sessionId)
	return f.Current(context.Background(), sessionId)
}

func (f *fakeAISessions) Current(_ context.Context, sessionId string) (*entity.AISessionContext, error) {
	current := entity.NewAISessionContext(sessionId, f.localId, "dev")
	return &current, nil
}

func (f *fakeAISessions) Enter(context.Context, string, string) (*entity.RegisteredWorkspace, error) {
	return &entity.RegisteredWorkspace{}, nil
}

func (f *fakeAISessions) ChooseEnvironment(context.Context, string, string) (*entity.AISessionContext, error) {
	return &entity.AISessionContext{}, nil
}

func (f *fakeAISessions) ResolveEnvironment(context.Context, string, string) (string, bool, error) {
	return "", true, nil
}

func (f *fakeAISessions) Close(_ context.Context, sessionId string) error {
	f.closed = append(f.closed, sessionId)
	return nil
}

func newTestMCPCall(active string) (adapter.MCPCallController, *fakeAIActivityRecorder, *deniedSessions) {
	call, recorder, approvals, _ := newTestMCPCallWithSessions(active)
	return call, recorder, approvals
}

func newTestMCPCallWithSessions(active string) (adapter.MCPCallController, *fakeAIActivityRecorder, *deniedSessions, *fakeAISessions) {
	recorder, approvals, sessions := &fakeAIActivityRecorder{}, &deniedSessions{}, &fakeAISessions{localId: active}
	call := NewMCPCall(NewErrorHandler(), sessions, maskingAIOutputGuard{}, recorder, silentAIChangeNotifier{}, approvals)
	return call, recorder, approvals, sessions
}

func TestMCPCallRejectsAnUnknownActionAndRecordsTheActivity(t *testing.T) {
	call, recorder, _ := newTestMCPCall("ws-1")
	tree := NewMCPTree(NewErrorHandler(), nil, nil, nil, nil, nil, nil, nil)
	outcome := call.Call(context.Background(), dto.NewMCPCall("session-1", "Claude Code", json.RawMessage(`{"action":"explode","note":"s3cret"}`)), tree)
	if !outcome.Failed || !strings.Contains(outcome.Text, `"code":"MCP-07400"`) || strings.Contains(outcome.Text, "s3cret") {
		t.Fatalf("Call() = %+v", outcome)
	}
	record := recorder.records[0]
	if record.client != "Claude Code" || record.activity.Action != "explode" || record.activity.Outcome != aiactivityoutcome.Error || record.activity.Code != "MCP-07400" || record.localId != "ws-1" || !strings.Contains(record.activity.Args, "s3cret") {
		t.Errorf("record = %+v", record)
	}
}

func TestMCPCallNeedsAnActiveWorkspaceForScopedTools(t *testing.T) {
	call, recorder, _ := newTestMCPCall("")
	tree := NewMCPTree(NewErrorHandler(), nil, nil, nil, nil, nil, nil, nil)
	scoped := call.Call(context.Background(), dto.NewMCPCall("session-1", "", json.RawMessage(`{"action":"list"}`)), tree)
	if !scoped.Failed || !strings.Contains(scoped.Text, "MCP-06404") {
		t.Fatalf("a scoped tool without a workspace = %+v", scoped)
	}
	workspace := NewMCPWorkspace(NewErrorHandler(), activeWorkspace{}, nil)
	unscoped := call.Call(context.Background(), dto.NewMCPCall("session-1", "", json.RawMessage(`{"action":3}`)), workspace)
	if !unscoped.Failed || !strings.Contains(unscoped.Text, "MCP-07400") {
		t.Errorf("a non-string action = %+v", unscoped)
	}
	if len(recorder.records) != 2 || recorder.records[0].client != entity.DefaultAIClientName {
		t.Errorf("records = %+v", recorder.records)
	}
}

func TestMCPCallClosingASessionDeniesItsApprovalsAndForgetsItsContext(t *testing.T) {
	call, _, approvals, sessions := newTestMCPCallWithSessions("ws-1")
	if err := call.CloseSession(context.Background(), dto.NewMCPSession("session-1")); err != nil {
		t.Fatal(err)
	}
	if len(approvals.ids) != 1 || approvals.ids[0] != "session-1" {
		t.Fatalf("denied sessions = %v", approvals.ids)
	}
	if len(sessions.closed) != 1 || sessions.closed[0] != "session-1" {
		t.Fatalf("closed sessions = %v", sessions.closed)
	}
}

func TestMCPCallOpeningASessionSeedsItsContext(t *testing.T) {
	call, _, _, sessions := newTestMCPCallWithSessions("ws-1")
	if err := call.OpenSession(context.Background(), dto.NewMCPSession("session-1")); err != nil {
		t.Fatal(err)
	}
	if len(sessions.opened) != 1 || sessions.opened[0] != "session-1" {
		t.Fatalf("opened sessions = %v", sessions.opened)
	}
}

func TestMCPCallRecordsTheSessionTag(t *testing.T) {
	call, recorder, _ := newTestMCPCall("ws-1")
	tree := NewMCPTree(NewErrorHandler(), nil, nil, nil, nil, nil, nil, nil)
	call.Call(context.Background(), dto.NewMCPCall("session-1", "Claude Code", json.RawMessage(`{"action":"explode"}`)), tree)
	if len(recorder.records) != 1 || recorder.records[0].session != "sessio" {
		t.Fatalf("records = %+v", recorder.records)
	}
}
