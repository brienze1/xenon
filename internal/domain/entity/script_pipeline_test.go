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
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

func entryWithID(id string) ConsoleEntry {
	return ConsoleEntry{ID: id}
}

func entryIDs(entries []ConsoleEntry) string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return strings.Join(ids, ",")
}

func TestConsoleBufferKeepsTheNewestEntries(t *testing.T) {
	buffer := NewConsoleBuffer().Appended([]ConsoleEntry{entryWithID("a"), entryWithID("b"), entryWithID("c")})
	if got := entryIDs(buffer.After("", 10)); got != "a,b,c" {
		t.Fatalf("all = %s", got)
	}
	if got := entryIDs(buffer.After("a", 1)); got != "b" {
		t.Errorf("after a, limit 1 = %s", got)
	}
	if got := entryIDs(buffer.After("c", 5)); got != "" {
		t.Errorf("after the newest = %s", got)
	}
	if got := entryIDs(buffer.After("gone", 2)); got != "a,b" {
		t.Errorf("unknown id starts from the oldest, got %s", got)
	}
	full := NewConsoleBuffer()
	for index := 0; index < ConsoleCapacity; index++ {
		full = full.Appended([]ConsoleEntry{entryWithID("old")})
	}
	full = full.Appended([]ConsoleEntry{entryWithID("new")})
	if len(full.Entries) != ConsoleCapacity || full.Entries[ConsoleCapacity-1].ID != "new" {
		t.Errorf("capacity = %d, last = %s", len(full.Entries), full.Entries[len(full.Entries)-1].ID)
	}
	listed := buffer.After("", 10)
	listed[0].ID = "changed"
	if buffer.Entries[0].ID != "a" {
		t.Errorf("After returns a copy")
	}
}

func TestConsoleErrorEntryAndMasking(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	scriptError := NewScriptError(scripterrorkind.Runtime, "Error: token s3cret", preRef()).At(2, 7)
	entry := NewConsoleErrorEntry("01E", "01S", scriptError, at)
	if entry.Kind != consoleentrykind.Error || entry.Level != consolelevel.Error || *entry.Script != preRef() || entry.Summary != "Error: token s3cret" || !reflect.DeepEqual(entry.Arguments, []string{"Error: token s3cret"}) || entry.ID != "01E" || entry.SendID != "01S" || !entry.CreatedAt.Equal(at) {
		t.Fatalf("entry = %+v", entry)
	}
	mask := func(text string) (string, bool) {
		masked := strings.ReplaceAll(text, "s3cret", MaskedValue)
		return masked, masked != text
	}
	masked := entry.WithSecretsMasked(mask)
	if !masked.Masked || masked.Summary != "Error: token "+MaskedValue || masked.Arguments[0] != "Error: token "+MaskedValue || entry.Arguments[0] != "Error: token s3cret" {
		t.Errorf("masked = %+v", masked)
	}
	onlyArgument := NewConsoleLogEntry("1", "s", consolelevel.Log, preRef(), "plain", []string{"plain", "s3cret"}, at).WithSecretsMasked(mask)
	if !onlyArgument.Masked || onlyArgument.Arguments[1] != MaskedValue {
		t.Errorf("an argument with a secret marks the entry masked: %+v", onlyArgument)
	}
	clean := NewConsoleLogEntry("1", "s", consolelevel.Log, preRef(), "plain", []string{"plain"}, at).WithSecretsMasked(mask)
	if clean.Masked {
		t.Errorf("nothing to mask: %+v", clean)
	}
	alreadyMasked := clean
	alreadyMasked.Masked = true
	if !alreadyMasked.WithSecretsMasked(mask).Masked {
		t.Errorf("a masked entry stays masked")
	}
}

func TestLocalVariablesAreTheStrongestScope(t *testing.T) {
	text := map[any]string{"x": "x", 42.0: "42", true: "true", nil: "null"}
	for value, want := range text {
		if got := LocalVariableText(value); got != want {
			t.Errorf("LocalVariableText(%v) = %q", value, got)
		}
	}
	if got := LocalVariableText(map[string]any{"a": 1.0}); got != `{"a":1}` {
		t.Errorf("object = %q", got)
	}
	base := NewVariableContext(NewVariableLayer(variablescope.Environment, []VariableRow{NewVariableRow("token", "env"), NewVariableRow("tenant", "acme")}, nil, nil, false))
	context := base.WithLocal(map[string]any{"token": "local", "page": 2.0})
	if got := context.Resolve("{{token}} {{tenant}} {{page}}", nil); got != "local acme 2" {
		t.Errorf("resolved = %q", got)
	}
	variable, _ := context.Lookup("token")
	if variable.Scope != variablescope.Local || len(base.Layers) != 1 || len(context.Layers) != 2 {
		t.Errorf("local layer = %+v, base layers %d", variable, len(base.Layers))
	}
}

func TestScriptOwnersFollowTheTree(t *testing.T) {
	collection := NewCollection("01C", "Billing")
	collection.Scripts = Scripts{PreRequest: "c"}
	folder := NewFolder("01F", "Invoices")
	folder.InheritScripts = false
	request := NewRequest("01R", "Create")
	if !collection.InheritScripts || !request.InheritScripts {
		t.Fatalf("inherit is on by default")
	}
	stored := NewRequest("01R", "Create")
	stored.Scripts = Scripts{PostResponse: "r"}
	stored.InheritScripts = false
	owners := request.WithScriptsOf(stored).ScriptOwners([]Collection{collection, folder})
	want := []ScriptOwner{
		NewScriptOwner(scriptownerkind.Collection, "01C", "Billing", Scripts{PreRequest: "c"}, true),
		NewScriptOwner(scriptownerkind.Folder, "01F", "Invoices", Scripts{}, false),
		NewScriptOwner(scriptownerkind.Request, "01R", "Create", Scripts{PostResponse: "r"}, false),
	}
	if !reflect.DeepEqual(owners, want) {
		t.Errorf("owners = %+v", owners)
	}
}

func TestSendExecutionCarriesScriptEffects(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	written := NewRequest("01R", "Create")
	execution := NewSendExecution("01S")
	execution.Start(written)
	context := execution.ScriptContext(VariableContext{})
	context.LocalVariables["leak"] = true
	if context.Request.ID != "01R" || context.Response != nil || len(execution.LocalVariables) != 0 {
		t.Fatalf("context = %+v, local = %v", context, execution.LocalVariables)
	}
	changed := written
	changed.Method = "POST"
	logs := []ScriptLog{{Level: consolelevel.Log, Arguments: []string{"a", "b"}}}
	nested := []NestedRequest{NewRejectedNestedRequest("locked")}
	pre := NewScriptRun(preRef(), start).WithEffects(changed, map[string]any{"x": "1"}, logs, nested)
	logs[0].Arguments[0] = "z"
	execution.AddRun(pre)
	if execution.Request.Method != "POST" || execution.LocalVariables["x"] != "1" || len(execution.Nested) != 1 || pre.Logs[0].Text() != "z b" {
		t.Fatalf("after pre = %+v", execution)
	}
	post := NewScriptRun(postRef(), start).WithEffects(written, map[string]any{"y": "2"}, nil, nil)
	execution.AddRun(post)
	if execution.Request.Method != "POST" || execution.LocalVariables["y"] != "2" || execution.LocalVariables["x"] != nil {
		t.Errorf("post-response runs keep the sent request and replace local variables: %+v", execution)
	}
	if got := execution.Variables(VariableContext{}).Resolve("{{y}}", nil); got != "2" {
		t.Errorf("variables = %q", got)
	}
	if execution.RunsPostScripts() {
		t.Errorf("no response yet")
	}
	execution.RecordResponse(EffectiveRequest{}, Response{StatusCode: 200})
	if !execution.RunsPostScripts() || execution.ScriptContext(VariableContext{}).Response.StatusCode != 200 {
		t.Errorf("a sent response runs post-response scripts")
	}
	execution.RecordResponse(EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Dns, "x"), Timings{}))
	if execution.RunsPostScripts() {
		t.Errorf("a failed send runs no post-response scripts")
	}
	if got := execution.LogValue(); !reflect.DeepEqual(got, map[string]any{"id": "01S", "outcome": sendoutcome.Error, "pre_runs": 1, "post_runs": 1, "status_code": 0}) {
		t.Errorf("LogValue = %v", got)
	}
	if got := NewSendExecution("x").LogValue(); !reflect.DeepEqual(got, map[string]any{"id": "x", "outcome": enums.SendOutcome(""), "pre_runs": 0, "post_runs": 0}) {
		t.Errorf("LogValue without response = %v", got)
	}
}

func TestHistoryEntriesOfScriptOutcomes(t *testing.T) {
	request := NewRequest("01R", "Create")
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stopped := NewSendExecution("s")
	stopped.AddRun(NewScriptRun(preRef(), at).Failed(NewScriptError(scripterrorkind.Runtime, "x", preRef())))
	entry := NewHistoryEntry("h", at, request, stopped)
	if entry.Outcome != sendoutcome.StoppedByScriptError || entry.WentOut() || entry.HasStatus() || entry.Result.StatusCode != 0 {
		t.Errorf("stopped = %+v", entry)
	}
	skipped := NewSendExecution("s")
	skipped.AddRun(NewScriptRun(preRef(), at).SkippedRequest())
	if NewHistoryEntry("h", at, request, skipped).WentOut() {
		t.Errorf("a skipped send did not go out")
	}
	sent := NewSendExecution("s")
	sent.RecordResponse(EffectiveRequest{}, Response{StatusCode: 201})
	if sentEntry := NewHistoryEntry("h", at, request, sent); !sentEntry.WentOut() || !sentEntry.HasStatus() || sentEntry.Result.StatusCode != 201 {
		t.Errorf("sent = %+v", sentEntry)
	}
	failed := NewSendExecution("s")
	failed.RecordResponse(EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Dns, "x"), Timings{}))
	if failedEntry := NewHistoryEntry("h", at, request, failed); !failedEntry.WentOut() || failedEntry.HasStatus() {
		t.Errorf("failed = %+v", failedEntry)
	}
}

func TestNestedRequestsAndPreflightMessages(t *testing.T) {
	sent := NewSentNestedRequest(EffectiveRequest{Method: "POST", URL: "http://x"}, Response{StatusCode: 200})
	if sent.HasFailed() || sent.FailureMessage() != "" || sent.Response.StatusCode != 200 {
		t.Errorf("sent = %+v", sent)
	}
	if got := sent.LogValue(); !reflect.DeepEqual(got, map[string]any{"method": "POST", "url": "http://x", "error": "", "status_code": 200}) {
		t.Errorf("LogValue = %v", got)
	}
	rejected := NewRejectedNestedRequest("locked")
	if !rejected.HasFailed() || rejected.FailureMessage() != "locked" || rejected.Response != nil {
		t.Errorf("rejected = %+v", rejected)
	}
	if got := rejected.LogValue(); !reflect.DeepEqual(got, map[string]any{"method": "", "url": "", "error": "locked"}) {
		t.Errorf("LogValue = %v", got)
	}
	unreachable := NewSentNestedRequest(EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Dns, "nowhere.test"), Timings{}))
	if !unreachable.HasFailed() || unreachable.FailureMessage() != unreachable.Response.Error.Message || unreachable.FailureMessage() == "" {
		t.Errorf("unreachable = %+v", unreachable)
	}
	preflight := SendPreflight{MissingSecrets: []string{"a", "b"}, LockedSecrets: []string{"c"}}
	if got := preflight.UnavailableSecretsMessage(); got != "The request uses secrets that have no value on this machine or are locked: a, b, c" {
		t.Errorf("message = %q", got)
	}
}

func TestAppSettingsFallBackToTheDefault(t *testing.T) {
	if got := (AppSettings{ScriptTimeoutMs: 2500}).OrDefault(); got.ScriptTimeoutMs != 2500 {
		t.Errorf("valid = %+v", got)
	}
	if got := (AppSettings{ScriptTimeoutMs: 5}).OrDefault(); got != DefaultAppSettings() {
		t.Errorf("invalid = %+v", got)
	}
	if got := (AppSettings{ScriptTimeoutMs: 5}).LogValue(); !reflect.DeepEqual(got, map[string]any{"script_timeout_ms": 5}) {
		t.Errorf("LogValue = %v", got)
	}
}

func TestScriptRunEffectsAreCopies(t *testing.T) {
	run := NewScriptRun(ScriptRef{Phase: scriptphase.PreRequest}, time.Time{})
	if run.Logs == nil || run.Nested == nil {
		t.Fatalf("a new run has empty logs and nested requests")
	}
	nested := []NestedRequest{{Error: "a"}}
	withEffects := run.WithEffects(Request{}, nil, nil, nested)
	nested[0].Error = "b"
	if withEffects.Nested[0].Error != "a" || withEffects.Logs == nil {
		t.Errorf("effects = %+v", withEffects)
	}
}
