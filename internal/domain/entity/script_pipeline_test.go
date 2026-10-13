package entity

import (
	"reflect"
	"slices"
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
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
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
	buffer, dropped := NewConsoleBuffer().Appended([]ConsoleEntry{entryWithID("a"), entryWithID("b"), entryWithID("c")})
	if got := entryIDs(buffer.Find(ConsoleQuery{Limit: 10})); got != "a,b,c" || len(dropped) != 0 {
		t.Fatalf("all = %s, dropped = %d", got, len(dropped))
	}
	if got := entryIDs(buffer.Find(ConsoleQuery{AfterID: "a", Limit: 1})); got != "b" {
		t.Errorf("after a, limit 1 = %s", got)
	}
	if got := entryIDs(buffer.Find(ConsoleQuery{AfterID: "c", Limit: 5})); got != "" {
		t.Errorf("after the newest = %s", got)
	}
	if got := entryIDs(buffer.Find(ConsoleQuery{AfterID: "gone", Limit: 2})); got != "a,b" {
		t.Errorf("unknown id starts from the oldest, got %s", got)
	}
	full := NewConsoleBuffer()
	for index := 0; index < ConsoleCapacity; index++ {
		full, _ = full.Appended([]ConsoleEntry{entryWithID("old")})
	}
	full, dropped = full.Appended([]ConsoleEntry{entryWithID("new"), entryWithID("newer")})
	if len(full.Entries) != ConsoleCapacity || full.Entries[ConsoleCapacity-1].ID != "newer" || full.Entries[ConsoleCapacity-2].ID != "new" || len(dropped) != 2 || dropped[0].ID != "old" {
		t.Errorf("capacity = %d, last = %s, dropped = %d", len(full.Entries), full.Entries[len(full.Entries)-1].ID, len(dropped))
	}
	if entry, found := buffer.Entry("b"); !found || entry.ID != "b" {
		t.Errorf("Entry(b) = %+v %v", entry, found)
	}
	if _, found := buffer.Entry("gone"); found {
		t.Errorf("Entry(gone) found")
	}
	listed := buffer.Find(ConsoleQuery{Limit: 10})
	listed[0].ID = "changed"
	if buffer.Entries[0].ID != "a" {
		t.Errorf("After returns a copy")
	}
}

func TestConsoleErrorEntryAndMasking(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	scriptError := NewScriptError(scripterrorkind.Runtime, "Error: token s3cret", preRef()).At(2, 7)
	entry := NewConsoleErrorEntry("01E", "01S", scriptError, at)
	if entry.Kind != consoleentrykind.Error || entry.Level != consolelevel.Error || *entry.Script != preRef() || entry.Summary != "Error: token s3cret" || entry.Line != 2 || entry.Column != 7 || !reflect.DeepEqual(entry.Arguments, []string{"Error: token s3cret"}) || entry.ID != "01E" || entry.SendID != "01S" || !entry.CreatedAt.Equal(at) {
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
	onlyArgument := NewConsoleLogEntry("1", "s", preRef(), NewScriptLog(consolelevel.Log, []string{"plain", "s3cret"}, nil), at)
	onlyArgument.Summary = "plain"
	onlyArgument = onlyArgument.WithSecretsMasked(mask)
	if !onlyArgument.Masked || onlyArgument.Arguments[1] != MaskedValue {
		t.Errorf("an argument with a secret marks the entry masked: %+v", onlyArgument)
	}
	clean := NewConsoleLogEntry("1", "s", preRef(), NewScriptLog(consolelevel.Log, []string{"plain"}, nil), at).WithSecretsMasked(mask)
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
	scopes := ScriptScopes{Environment: NewScriptScope("dev", []ScriptVariable{NewScriptVariable("token", "t")}), Local: NewScriptScope("", nil)}
	execution.Start(written, scopes, map[string]string{"key": "k"})
	context := execution.ScriptContext(ScriptContext{Info: NewScriptInfo("01R", "Create", "Create", []string{"Create"})})
	context.Vault["leak"] = "x"
	if context.Request.ID != "01R" || context.Response != nil || len(execution.Vault) != 1 || context.Scopes.Environment.Name != "dev" || context.Info.RequestName != "Create" {
		t.Fatalf("context = %+v, vault = %v", context, execution.Vault)
	}
	changed := written
	changed.Method = "POST"
	logs := []ScriptLog{NewScriptLog(consolelevel.Log, []string{"a", "b"}, []any{"a", "b"})}
	nested := []NestedRequest{NewRejectedNestedRequest("locked")}
	tests := []TestResult{{Name: "t", Status: teststatus.Passed}}
	pre := NewScriptRun(preRef(), start).WithEffects(ScriptEffects{Request: changed, Scopes: scopes.WithLocal(NewScriptScope("", []ScriptVariable{NewScriptVariable("x", "1")})), Vault: map[string]string{"key": "k2"}, Logs: logs, Nested: nested, Tests: tests})
	logs[0].Arguments[0] = "z"
	execution.AddRun(pre)
	if execution.Request.Method != "POST" || execution.Scopes.Local.ValueMap()["x"] != "1" || len(execution.Nested) != 1 || pre.Logs[0].Text() != "z b" || execution.Vault["key"] != "k2" {
		t.Fatalf("after pre = %+v", execution)
	}
	if len(execution.Tests) != 1 || execution.Tests[0].Script != preRef() || pre.TestCount != 1 {
		t.Errorf("tests = %+v", execution.Tests)
	}
	post := NewScriptRun(postRef(), start).WithEffects(ScriptEffects{Request: written, Scopes: scopes.WithLocal(NewScriptScope("", []ScriptVariable{NewScriptVariable("y", "2")}))})
	execution.AddRun(post)
	if execution.Request.Method != "POST" || execution.Scopes.Local.ValueMap()["y"] != "2" || execution.Scopes.Local.ValueMap()["x"] != nil {
		t.Errorf("post-response runs keep the sent request and replace the scopes: %+v", execution)
	}
	if got := execution.Variables(VariableContext{}).Resolve("{{y}}", nil); got != "2" {
		t.Errorf("variables = %q", got)
	}
	if execution.RunsPostScripts() {
		t.Errorf("no response yet")
	}
	execution.RecordResponse(EffectiveRequest{}, Response{StatusCode: 200})
	if !execution.RunsPostScripts() || execution.ScriptContext(ScriptContext{}).Response.StatusCode != 200 {
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
	entry := NewHistoryEntry("h", at, request, stopped, UserSendOrigin())
	if entry.Outcome != sendoutcome.StoppedByScriptError || entry.WentOut() || entry.HasStatus() || entry.Result.StatusCode != 0 {
		t.Errorf("stopped = %+v", entry)
	}
	skipped := NewSendExecution("s")
	skipped.AddRun(NewScriptRun(preRef(), at).SkippedRequest())
	if NewHistoryEntry("h", at, request, skipped, UserSendOrigin()).WentOut() {
		t.Errorf("a skipped send did not go out")
	}
	sent := NewSendExecution("s")
	sent.RecordResponse(EffectiveRequest{}, Response{StatusCode: 201})
	if sentEntry := NewHistoryEntry("h", at, request, sent, UserSendOrigin()); !sentEntry.WentOut() || !sentEntry.HasStatus() || sentEntry.Result.StatusCode != 201 {
		t.Errorf("sent = %+v", sentEntry)
	}
	failed := NewSendExecution("s")
	failed.RecordResponse(EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Dns, "x"), Timings{}))
	if failedEntry := NewHistoryEntry("h", at, request, failed, UserSendOrigin()); !failedEntry.WentOut() || failedEntry.HasStatus() {
		t.Errorf("failed = %+v", failedEntry)
	}
}

func TestNestedRequestsAndPreflightMessages(t *testing.T) {
	sent := NewSentNestedRequest(EffectiveRequest{Method: "POST", URL: "http://x"}, Response{StatusCode: 200})
	if sent.HasFailed() || sent.FailureMessage() != "" || sent.Response.StatusCode != 200 {
		t.Errorf("sent = %+v", sent)
	}
	if got := sent.LogValue(); !reflect.DeepEqual(got, map[string]any{"method": "POST", "url": "http://x", "error": "", "requestId": "", "skipped": false, "status_code": 200}) {
		t.Errorf("LogValue = %v", got)
	}
	rejected := NewRejectedNestedRequest("locked")
	if !rejected.HasFailed() || rejected.FailureMessage() != "locked" || rejected.Response != nil {
		t.Errorf("rejected = %+v", rejected)
	}
	if got := rejected.LogValue(); !reflect.DeepEqual(got, map[string]any{"method": "", "url": "", "error": "locked", "requestId": "", "skipped": false}) {
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
	if got := (AppSettings{ScriptTimeoutMs: 5, AIAccess: DefaultAIAccessSettings()}).OrDefault(); got != DefaultAppSettings() {
		t.Errorf("invalid = %+v", got)
	}
	if got := (AppSettings{ScriptTimeoutMs: 5}).LogValue().(map[string]any); got["script_timeout_ms"] != 5 {
		t.Errorf("LogValue = %v", got)
	}
}

func TestScriptRunEffectsAreCopies(t *testing.T) {
	run := NewScriptRun(ScriptRef{Phase: scriptphase.PreRequest}, time.Time{})
	if run.Logs == nil || run.Nested == nil {
		t.Fatalf("a new run has empty logs and nested requests")
	}
	nested := []NestedRequest{{Error: "a"}}
	withEffects := run.WithEffects(ScriptEffects{Nested: nested})
	nested[0].Error = "b"
	if withEffects.Nested[0].Error != "a" || withEffects.Logs == nil {
		t.Errorf("effects = %+v", withEffects)
	}
}

func TestScriptScopeChangesAndLookups(t *testing.T) {
	before := NewScriptScope("dev", []ScriptVariable{NewScriptVariable("a", "1"), NewScriptVariable("b", "2"), {Key: "off", Value: "x", Disabled: true}})
	after := before.WithValue("a", 1.0).WithValue("c", map[string]any{"x": true})
	after.Values = append(after.Values[:1], after.Values[2:]...)
	set, unset := after.ChangesFrom(before)
	if !reflect.DeepEqual(set, []KeyValue{NewKeyValue("c", `{"x":true}`)}) || !reflect.DeepEqual(unset, []string{"b"}) {
		t.Errorf("changes = %+v %v", set, unset)
	}
	if variable, found := after.Lookup("c"); !found || variable.Text() != `{"x":true}` {
		t.Errorf("lookup = %+v %v", variable, found)
	}
	if _, found := before.Lookup("off"); found {
		t.Errorf("disabled values are not found")
	}
	if NewInactiveScriptScope().Inactive != true || (ScriptScopes{Globals: before}).Scope(variablescope.Globals).Name != "dev" {
		t.Errorf("scopes")
	}
	for value, want := range map[any]string{"s": "s", 2.0: "2", nil: "null"} {
		if got := LocalVariableText(value); got != want {
			t.Errorf("LocalVariableText(%v) = %q", value, got)
		}
	}
	tagged := map[string]any{"$type": "undefined"}
	if LocalVariableText(tagged) != "" || LocalVariableText(map[string]any{"$type": "number", "value": "NaN"}) != "NaN" || LocalVariableText(map[string]any{"$type": "map", "value": []any{}}) != "[]" {
		t.Errorf("tagged values read as their value")
	}
}

func TestVariableContextWithScriptScopesAndVault(t *testing.T) {
	rows := []VariableRow{NewVariableRow("base", "https://a"), NewVariableRow("gone", "x"), {Key: "missing", Secret: true, Enabled: true}}
	context := NewVariableContext(NewVariableLayer(variablescope.Environment, rows, nil, nil, false), NewVariableLayer(variablescope.Globals, []VariableRow{{Key: "key", Secret: true, Enabled: true}}, nil, map[string]string{"key": "s3"}, false))
	scopes := ScriptScopes{
		Environment: NewScriptScope("dev", []ScriptVariable{NewScriptVariable("base", "https://b"), NewScriptVariable("new", 1.0)}),
		Globals:     NewScriptScope("", []ScriptVariable{{Key: "key", Value: "s3", Secret: true}}),
		Local:       NewScriptScope("", []ScriptVariable{NewScriptVariable("id", "7")}),
	}
	updated := context.WithScriptScopes(scopes)
	if got := updated.Resolve("{{base}}/{{new}}/{{id}}/{{gone}}/{{vault:key}}", nil); got != "https://b/1/7/{{gone}}/s3" {
		t.Errorf("resolve = %q", got)
	}
	if missing := updated.UnavailableSecretsIn("{{missing}}"); len(missing) != 1 {
		t.Errorf("unavailable secrets survive script scopes: %+v", missing)
	}
	vaulted := updated.WithVault(map[string]string{"key": "s4"})
	if got := vaulted.Resolve("{{vault:key}}/{{vault:nope}}", nil); got != "s4/{{vault:nope}}" || !slices.Contains(vaulted.SecretValues(), "s4") {
		t.Errorf("vault resolve = %q", got)
	}
	if got := context.WithScriptScopes(ScriptScopes{}).Resolve("{{base}}", nil); got != "https://a" {
		t.Errorf("scopes never loaded leave the context as it is: %q", got)
	}
}

func TestVariableTableWithScriptChanges(t *testing.T) {
	rows := []VariableRow{NewVariableRow("plain", "init"), {Key: "secret", Secret: true, Enabled: true}}
	table := VariableTable{Scope: NewGlobalsScopeRef(), Rows: rows, Current: map[string]string{"plain": "cur"}, Secrets: map[string]string{"secret": "old"}}
	changed := table.WithScriptChanges([]KeyValue{NewKeyValue("plain", "p2"), NewKeyValue("secret", "s2"), NewKeyValue("added", "a")}, []string{"nope"}, false)
	if changed.Current["plain"] != "p2" || changed.Secrets["secret"] != "s2" || len(changed.Rows) != 3 || changed.Rows[2].InitialValue != "" || changed.Current["added"] != "a" {
		t.Errorf("set = %+v", changed)
	}
	unset := table.WithScriptChanges(nil, []string{"plain", "secret"}, false)
	if _, found := unset.Current["plain"]; found || unset.Secrets["secret"] != "old" {
		t.Errorf("unset drops plain current values only = %+v", unset)
	}
	vault := table.WithScriptChanges([]KeyValue{NewKeyValue("plain", "x"), NewKeyValue("fresh", "f")}, []string{"secret"}, true)
	if vault.Current["plain"] != "cur" || !vault.Rows[2].Secret || vault.Secrets["fresh"] != "f" {
		t.Errorf("vault changes only touch secret rows = %+v", vault)
	}
	if _, found := vault.Secrets["secret"]; found {
		t.Errorf("vault unset drops the secret value")
	}
}

func TestScriptCookiesFromResponseHeaders(t *testing.T) {
	headers := []KeyValue{NewKeyValue("Set-Cookie", "session=s1; Path=/; HttpOnly; Secure; Max-Age=60"), NewKeyValue("set-cookie", "theme=dark; Domain=.example.com; Expires=Wed, 21 Oct 2026 07:28:00 GMT"), NewKeyValue("Set-Cookie", "=bad"), NewKeyValue("X-Other", "a=b")}
	cookies := NewScriptCookies(headers, "api.example.com")
	if len(cookies) != 2 {
		t.Fatalf("cookies = %+v", cookies)
	}
	if cookies[0] != (ScriptCookie{Name: "session", Value: "s1", Domain: "api.example.com", Path: "/", MaxAge: 60, Secure: true, HttpOnly: true, HostOnly: true}) {
		t.Errorf("session = %+v", cookies[0])
	}
	if cookies[1].Domain != "example.com" || cookies[1].HostOnly || cookies[1].Expires != "Wed, 21 Oct 2026 07:28:00 UTC" {
		t.Errorf("theme = %+v", cookies[1])
	}
	if (EffectiveRequest{URL: "https://api.example.com:8443/x"}).Host() != "api.example.com" || (EffectiveRequest{URL: "%%"}).Host() != "" {
		t.Errorf("host")
	}
}

func TestRunNestedRequests(t *testing.T) {
	execution := NewSendExecution("x")
	execution.Start(NewRequest("01R", "R"), ScriptScopes{}, nil)
	sent := EffectiveRequest{URL: "https://a"}
	execution.RecordResponse(sent, Response{StatusCode: 200})
	nested := NewRunNestedRequest("01R", execution)
	if !nested.IsRunRequest() || nested.Skipped || nested.Response.StatusCode != 200 || nested.Request.URL != "https://a" || nested.HasFailed() {
		t.Errorf("run = %+v", nested)
	}
	skipped := NewSendExecution("y")
	skipped.AddRun(NewScriptRun(preRef(), time.Time{}).SkippedRequest())
	if run := NewRunNestedRequest("01R", skipped); !run.Skipped || run.Response != nil {
		t.Errorf("skipped run = %+v", run)
	}
	missing := NewMissingRunNestedRequest("01X", "nope")
	if !missing.HasFailed() || missing.FailureMessage() != "nope" || missing.LogValue().(map[string]any)["requestId"] != "01X" {
		t.Errorf("missing = %+v", missing)
	}
	if info := NewScriptInfo("01R", "R", "C", []string{"C", "R"}); !info.HasIteration() || (ScriptInfo{}).HasIteration() {
		t.Errorf("info = %+v", info)
	}
	if NewUnsavedVariablesError(preRef(), "locked").Message != "The variable changes of this script could not be saved: locked" {
		t.Errorf("unsaved")
	}
	if !NewScriptLog(consolelevel.Clear, nil, nil).IsClear() {
		t.Errorf("clear")
	}
}
