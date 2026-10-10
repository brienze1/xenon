package entity

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

func preRef() ScriptRef {
	return ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Request, OwnerName: "Login"}
}

func postRef() ScriptRef {
	return ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Folder, OwnerName: "Auth"}
}

func TestScriptRefLabel(t *testing.T) {
	if got := preRef().Label(); got != "Login · Pre-request" {
		t.Errorf("pre label = %q", got)
	}
	if got := postRef().Label(); got != "Auth · Post-response" {
		t.Errorf("post label = %q", got)
	}
	if !preRef().IsPreRequest() || postRef().IsPreRequest() {
		t.Errorf("IsPreRequest is wrong")
	}
}

func TestScriptOwner(t *testing.T) {
	scripts := Scripts{PreRequest: "pm.x()", PostResponse: "  ", Disabled: nil}
	owner := NewScriptOwner(scriptownerkind.Folder, "01F", "Auth", scripts, false)
	if !owner.CutsInheritance() {
		t.Errorf("a folder with inherit off cuts inheritance")
	}
	if NewScriptOwner(scriptownerkind.Folder, "01F", "Auth", scripts, true).CutsInheritance() {
		t.Errorf("a folder with inherit on does not cut inheritance")
	}
	if NewScriptOwner(scriptownerkind.Collection, "01C", "C", scripts, false).CutsInheritance() {
		t.Errorf("a collection never cuts inheritance")
	}
	if !owner.Runs(scriptphase.PreRequest) || owner.Runs(scriptphase.PostResponse) {
		t.Errorf("Runs follows script text")
	}
	disabled := NewScriptOwner(scriptownerkind.Request, "01R", "R", Scripts{PreRequest: "x", Disabled: []enums.ScriptPhase{scriptphase.PreRequest}}, true)
	if disabled.Runs(scriptphase.PreRequest) {
		t.Errorf("a disabled script does not run")
	}
	want := ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Folder, ItemID: "01F", OwnerName: "Auth", Text: "pm.x()"}
	if got := owner.Ref(scriptphase.PreRequest); got != want {
		t.Errorf("Ref = %+v", got)
	}
}

func TestScriptError(t *testing.T) {
	scriptError := NewScriptError(scripterrorkind.Runtime, "boom", preRef())
	if scriptError.HasLocation() || scriptError.Location() != "" {
		t.Errorf("no location expected: %+v", scriptError)
	}
	located := scriptError.At(3, 7)
	if !located.HasLocation() || located.Location() != "line 3, column 7" || located.Line != 3 || located.Column != 7 {
		t.Errorf("located = %+v (%q)", located, located.Location())
	}
	if NewScriptError(scripterrorkind.Syntax, "x", preRef()).At(1, 0).Location() != "line 1, column 0" {
		t.Errorf("line 1 is a location")
	}
	if got := located.WithMessage("masked"); got.Message != "masked" || got.Kind != scripterrorkind.Runtime || got.Script != preRef() {
		t.Errorf("WithMessage = %+v", got)
	}
}

func TestScriptRun(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	run := NewScriptRun(preRef(), start)
	if run.Status != scriptrunstatus.Ok || run.HasFailed() || run.HasSkippedRequest() || run.StopsSend() {
		t.Errorf("new run = %+v", run)
	}
	if got := run.Finished(start.Add(42 * time.Millisecond)).Duration; got != 42*time.Millisecond {
		t.Errorf("duration = %v", got)
	}
	failed := run.Failed(NewScriptError(scripterrorkind.Timeout, "t", preRef()))
	if !failed.HasFailed() || failed.Error == nil || failed.Error.Kind != scripterrorkind.Timeout || !failed.StopsSend() {
		t.Errorf("failed = %+v", failed)
	}
	skipped := run.SkippedRequest()
	if !skipped.HasSkippedRequest() || skipped.HasFailed() || !skipped.StopsSend() {
		t.Errorf("skipped = %+v", skipped)
	}
	postFailed := NewScriptRun(postRef(), start).Failed(NewScriptError(scripterrorkind.Runtime, "x", postRef()))
	if postFailed.StopsSend() || NewScriptRun(postRef(), start).SkippedRequest().StopsSend() {
		t.Errorf("post-response runs never stop the send")
	}
}

func TestTestResultStatus(t *testing.T) {
	cases := []struct {
		result                  TestResult
		passed, failed, skipped bool
	}{
		{TestResult{Status: teststatus.Passed}, true, false, false},
		{TestResult{Status: teststatus.Failed}, false, true, false},
		{TestResult{Status: teststatus.Skipped}, false, false, true},
	}
	for _, tc := range cases {
		if tc.result.IsPassed() != tc.passed || tc.result.IsFailed() != tc.failed || tc.result.IsSkipped() != tc.skipped {
			t.Errorf("status predicates wrong for %q", tc.result.Status)
		}
	}
}

func TestNestedRequest(t *testing.T) {
	if (NestedRequest{}).HasFailed() || (NestedRequest{Response: &Response{}}).HasFailed() {
		t.Errorf("a nested request without errors has not failed")
	}
	if !(NestedRequest{Error: "locked"}).HasFailed() || !(NestedRequest{Response: &Response{Error: &SendError{}}}).HasFailed() {
		t.Errorf("errors mark a nested request failed")
	}
	if !(NestedRequest{RequestID: "01R"}).IsRunRequest() || (NestedRequest{}).IsRunRequest() {
		t.Errorf("IsRunRequest follows the request id")
	}
}

func TestConsoleEntry(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	arguments := []string{`"a"`}
	entry := NewConsoleLogEntry("01E", "01S", consolelevel.Warn, preRef(), "hello", arguments, at)
	arguments[0] = "changed"
	want := ConsoleEntry{ID: "01E", SendID: "01S", Kind: consoleentrykind.Log, Level: consolelevel.Warn, Summary: "hello", Arguments: []string{`"a"`}, CreatedAt: at}
	if entry.Script == nil || *entry.Script != preRef() {
		t.Fatalf("script = %+v", entry.Script)
	}
	entry.Script = nil
	if !reflect.DeepEqual(entry, want) {
		t.Errorf("entry = %+v", entry)
	}
	long := strings.Repeat("é", 201)
	if got := ConsoleSummary(long); got != strings.Repeat("é", 200) {
		t.Errorf("summary is cut at 200 runes, got %d", len([]rune(got)))
	}
	if got := ConsoleSummary(strings.Repeat("é", 200)); got != strings.Repeat("é", 200) {
		t.Errorf("200 runes are kept")
	}
	kinds := []struct {
		kind                 string
		request, log, failed bool
	}{
		{"request", true, false, false},
		{"log", false, true, false},
		{"info", false, true, false},
		{"error", false, false, true},
		{"test_failure", false, false, true},
	}
	for _, tc := range kinds {
		e := ConsoleEntry{Kind: consoleentrykind.GetConsoleEntryKind(tc.kind)}
		if e.IsRequest() != tc.request || e.IsLog() != tc.log || e.IsError() != tc.failed {
			t.Errorf("predicates wrong for %s", tc.kind)
		}
	}
}

func TestAppSettings(t *testing.T) {
	settings := DefaultAppSettings()
	if settings.ScriptTimeoutMs != 10000 || settings.ScriptTimeout() != 10*time.Second || !settings.HasValidScriptTimeout() {
		t.Errorf("defaults = %+v", settings)
	}
	for timeout, valid := range map[int]bool{999: false, 1000: true, 60000: true, 60001: false} {
		if got := (AppSettings{ScriptTimeoutMs: timeout}).HasValidScriptTimeout(); got != valid {
			t.Errorf("HasValidScriptTimeout(%d) = %v", timeout, got)
		}
	}
}

func TestSendExecutionOutcomes(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	execution := NewSendExecution("01S")
	if execution.ID != "01S" || len(execution.Vault) != 0 || execution.Vault == nil || !execution.ShouldSend() {
		t.Fatalf("new execution = %+v", execution)
	}
	execution.AddRun(NewScriptRun(preRef(), start))
	execution.AddRun(NewScriptRun(postRef(), start).Failed(NewScriptError(scripterrorkind.Runtime, "post", postRef())))
	if len(execution.PreRuns) != 1 || len(execution.PostRuns) != 1 || !execution.ShouldSend() || execution.Outcome != "" {
		t.Errorf("runs = %+v", execution)
	}
	stopped := NewSendExecution("x")
	stopped.AddRun(NewScriptRun(preRef(), start).Failed(NewScriptError(scripterrorkind.Syntax, "pre", preRef())))
	if stopped.ShouldSend() || stopped.Outcome != sendoutcome.StoppedByScriptError {
		t.Errorf("stopped = %+v", stopped)
	}
	skipped := NewSendExecution("x")
	skipped.AddRun(NewScriptRun(preRef(), start).SkippedRequest())
	if skipped.ShouldSend() || skipped.Outcome != sendoutcome.SkippedByScript {
		t.Errorf("skipped = %+v", skipped)
	}
	postSkip := NewSendExecution("x")
	postSkip.AddRun(NewScriptRun(postRef(), start).SkippedRequest())
	if postSkip.Outcome != "" {
		t.Errorf("post skip outcome = %q", postSkip.Outcome)
	}
	if errs := execution.ScriptErrors(); len(errs) != 1 || errs[0].Message != "post" {
		t.Errorf("script errors = %+v", errs)
	}
	if errs := stopped.ScriptErrors(); len(errs) != 1 || errs[0].Message != "pre" {
		t.Errorf("pre errors = %+v", errs)
	}
}

func TestSendExecutionResponseTestsAndNested(t *testing.T) {
	execution := NewSendExecution("01S")
	sent := EffectiveRequest{Method: "GET", URL: "http://x"}
	execution.RecordResponse(sent, Response{StatusCode: 200})
	if execution.Outcome != sendoutcome.Sent || execution.SentRequest.URL != "http://x" || execution.Response.StatusCode != 200 {
		t.Errorf("sent = %+v", execution)
	}
	execution.RecordResponse(sent, Response{Error: &SendError{Kind: senderrorkind.Dns}})
	if execution.Outcome != sendoutcome.Error {
		t.Errorf("error outcome = %q", execution.Outcome)
	}
	execution.RecordResponse(sent, Response{Error: &SendError{Kind: senderrorkind.Cancelled}})
	if execution.Outcome != sendoutcome.Cancelled {
		t.Errorf("cancelled outcome = %q", execution.Outcome)
	}
	execution.AddTests([]TestResult{{Name: "a", Status: teststatus.Passed, Index: 9}, {Name: "b", Status: teststatus.Failed}})
	execution.AddTests([]TestResult{{Name: "c", Status: teststatus.Passed}})
	if len(execution.Tests) != 3 || execution.Tests[0].Index != 0 || execution.Tests[2].Index != 2 || execution.Tests[2].Name != "c" {
		t.Errorf("tests = %+v", execution.Tests)
	}
	if got := execution.PassedTestCount(); got != 2 {
		t.Errorf("passed = %d", got)
	}
	execution.AddNested(NestedRequest{RequestID: "01R"})
	if len(execution.Nested) != 1 || execution.Nested[0].RequestID != "01R" {
		t.Errorf("nested = %+v", execution.Nested)
	}
}
