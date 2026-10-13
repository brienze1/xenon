package sandbox

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

type fakeRequests struct {
	reply    func(entity.Request) entity.NestedRequest
	run      func(string, entity.ScriptScopes) entity.NestedRequest
	received []entity.Request
	scopes   []entity.ScriptScopes
	mutex    sync.Mutex
}

func (f *fakeRequests) Send(_ context.Context, request entity.Request, scopes entity.ScriptScopes) entity.NestedRequest {
	f.mutex.Lock()
	f.received = append(f.received, request)
	f.scopes = append(f.scopes, scopes)
	f.mutex.Unlock()
	return f.reply(request)
}

func (f *fakeRequests) Run(_ context.Context, requestId string, scopes entity.ScriptScopes) entity.NestedRequest {
	f.mutex.Lock()
	f.scopes = append(f.scopes, scopes)
	f.mutex.Unlock()
	return f.run(requestId, scopes)
}

func scopesOf(local map[string]any) entity.ScriptScopes {
	values := make([]entity.ScriptVariable, 0, len(local))
	for _, key := range []string{"kept", "gone", "before", "a", "k"} {
		if value, found := local[key]; found {
			values = append(values, entity.NewScriptVariable(key, value))
		}
	}
	return entity.ScriptScopes{
		Environment: entity.NewScriptScope("dev", []entity.ScriptVariable{entity.NewScriptVariable("tenant", "acme")}),
		Collection:  entity.NewScriptScope("", []entity.ScriptVariable{entity.NewScriptVariable("page", "1")}),
		Globals:     entity.NewScriptScope("", []entity.ScriptVariable{{Key: "secret", Value: "s", Secret: true}}),
		Local:       entity.NewScriptScope("", values),
	}
}

func contextOf(request entity.Request, local map[string]any) entity.ScriptContext {
	return entity.ScriptContext{Request: request, Scopes: scopesOf(local)}
}

// Windows runners load the Intl polyfill slowly while the whole module is tested in parallel.
const scriptTestTimeout = 20 * time.Second

func runIn(t *testing.T, phase entity.ScriptRef, scope entity.ScriptContext, requests usecase.SendScriptRequest, text string) entity.ScriptRun {
	t.Helper()
	phase.Text = text
	return testRunner(scriptTestTimeout, memoryGrowthLimit).Run(context.Background(), phase, scope, requests, nil, entity.DefaultAppSettings())
}

func postScript() entity.ScriptRef {
	post := script("")
	post.Phase = scriptphase.PostResponse
	return post
}

func ok(t *testing.T, run entity.ScriptRun) entity.ScriptRun {
	t.Helper()
	if run.HasFailed() {
		t.Fatalf("script failed: %+v", *run.Error)
	}
	return run
}

func logTexts(run entity.ScriptRun) string {
	texts := make([]string, 0, len(run.Logs))
	for _, log := range run.Logs {
		texts = append(texts, log.Text())
	}
	return strings.Join(texts, "|")
}

func TestPmVariableScopes(t *testing.T) {
	run := ok(t, runIn(t, script(""), contextOf(entity.NewRequest("", ""), map[string]any{"kept": "yes", "gone": 1.0}), nil, `
pm.variables.set("n", 2)
pm.variables.set("obj", { a: [1] })
pm.variables.set("nan", NaN)
pm.variables.unset("gone")
pm.environment.set("tenant", "globex")
pm.collectionVariables.unset("page")
pm.globals.set("count", 3)
console.log(pm.variables.get("kept"), pm.variables.get("n") + 1, pm.variables.get("obj").a[0], pm.variables.get("gone"), pm.variables.get("tenant"), pm.environment.name)
console.log(pm.variables.replaceIn("{{tenant}}/{{n}}/{{kept}}/{{unknown}}/{{secret}}"), pm.globals.get("secret"))`))
	if got := logTexts(run); got != "yes 3 1 undefined globex dev|globex/2/yes/{{unknown}}/s s" {
		t.Errorf("logs = %s", got)
	}
	local := run.Scopes.Local.ValueMap()
	if !reflect.DeepEqual(local, map[string]any{"kept": "yes", "n": 2.0, "obj": map[string]any{"a": []any{1.0}}, "nan": nil}) {
		t.Errorf("local = %#v", local)
	}
	if run.Scopes.Environment.ValueMap()["tenant"] != "globex" || run.Scopes.Environment.Name != "dev" || len(run.Scopes.Collection.Values) != 0 || run.Scopes.Globals.ValueMap()["count"] != 3.0 {
		t.Errorf("scopes = %+v", run.Scopes)
	}
	if secret, found := run.Scopes.Globals.Lookup("secret"); !found || !secret.Secret {
		t.Errorf("secret flags survive: %+v", secret)
	}
}

func TestPmScopesWithoutAnEnvironment(t *testing.T) {
	scopes := scopesOf(nil)
	scopes.Environment = entity.NewInactiveScriptScope()
	run := ok(t, runIn(t, script(""), entity.ScriptContext{Scopes: scopes}, nil, `
pm.environment.set("token", "t")
pm.environment.set("token", "t")
pm.environment.unset("token")
pm.environment.clear()
console.log(pm.environment.name, pm.environment.get("token"))`))
	if got := logTexts(run); got != "undefined undefined" {
		t.Errorf("logs = %s", got)
	}
	notices := make([]string, 0)
	for _, notice := range run.Notices {
		notices = append(notices, notice.Level.String()+" "+notice.Text())
	}
	want := []string{
		`warn pm.environment.set("token") was ignored: no environment is active`,
		`warn pm.environment.unset("token") was ignored: no environment is active`,
		`warn pm.environment.clear() was ignored: no environment is active`,
	}
	if !reflect.DeepEqual(notices, want) || !run.Scopes.Environment.Inactive || len(run.Scopes.Environment.Values) != 0 {
		t.Errorf("notices = %v, scopes = %+v", notices, run.Scopes.Environment)
	}
}

func TestConsoleCallsReachTheSinkInOrder(t *testing.T) {
	received := make([]string, 0)
	scope := entity.ScriptContext{Console: func(log entity.ScriptLog) { received = append(received, log.Level.String()+":"+log.Text()) }}
	run := ok(t, runIn(t, script(""), scope, nil, `
console.log("a", 1, { b: [true] })
pm.visualizer.set("<b></b>")
pm.visualizer.set("<b></b>")
console.clear()
setTimeout(() => console.warn("late"), 1)`))
	want := []string{`log:a 1 {"b":[true]}`, "info:pm.visualizer is not rendered in Xenon", "clear:", "warn:late"}
	if !reflect.DeepEqual(received, want) {
		t.Errorf("received = %v", received)
	}
	if len(run.Logs) != 3 || !run.Logs[1].IsClear() || len(run.Notices) != 1 {
		t.Errorf("logs = %+v notices = %+v", run.Logs, run.Notices)
	}
	if !reflect.DeepEqual(run.Logs[0].Values, []any{"a", 1.0, map[string]any{"b": []any{true}}}) {
		t.Errorf("values = %#v", run.Logs[0].Values)
	}
}

func TestConsoleValuesAreSerialisedAsPostmanSandboxDoes(t *testing.T) {
	run := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `
const cycle = { a: 1 }
cycle.self = cycle
console.log(undefined, NaN, -0, 10n, Symbol("s"), new Date(0), /x/g, new Map([["k", 1]]), new Set([1]), new Uint8Array([1, 2]), function named() {}, () => {}, new Error("e"), cycle, Buffer.from("hi"), new WeakMap(), new WeakSet(), new ArrayBuffer(2), [undefined])`))
	want := []any{
		map[string]any{"$type": "undefined"}, map[string]any{"$type": "number", "value": "NaN"}, 0.0, map[string]any{"$type": "bigint", "value": "10"},
		map[string]any{"$type": "symbol", "value": "Symbol(s)"}, map[string]any{"$type": "date", "value": "1970-01-01T00:00:00.000Z"}, map[string]any{},
		map[string]any{"$type": "map", "value": []any{[]any{"k", 1.0}}}, map[string]any{"$type": "set", "value": []any{1.0}}, map[string]any{"0": 1.0, "1": 2.0},
		"[Function: named]", "[Function]", map[string]any{"type": "Error", "name": "Error", "message": "e"}, map[string]any{"a": 1.0, "self": map[string]any{"$type": "circular"}},
		map[string]any{"0": 104.0, "1": 105.0}, "[WeakMap]", "[WeakSet]", "[ArrayBuffer { byteLength: 2 }]", []any{map[string]any{"$type": "undefined"}},
	}
	if !reflect.DeepEqual(run.Logs[0].Values, want) {
		t.Errorf("values =\n%#v\nwant\n%#v", run.Logs[0].Values, want)
	}
	texts := run.Logs[0].Arguments
	if texts[0] != "undefined" || texts[10] != "[Function: named]" || texts[12] != "Error: e" || !strings.Contains(texts[13], `"$type":"circular"`) || texts[7] != "{}" {
		t.Errorf("texts = %#v", texts)
	}
}

func TestPmRequestChangesComeBackToTheSend(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test/items?a=1#top"
	written.Headers = []entity.KeyValue{entity.NewKeyValue("X-Old", "1"), entity.NewKeyValue("Accept", "text/plain")}
	written.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Json, Content: "{}"}}
	written.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "t1"}}
	run := ok(t, runIn(t, script(""), contextOf(written, nil), nil, `
const url = pm.request.url
console.log(url.getHost(), url.getPath(), url.getQueryString(), url.getPathWithQuery(), url.getRemote(), url.protocol, url.hash, pm.request.auth.parameters().get("token"))
pm.request.method = "put"
pm.request.headers.upsert({ key: "accept", value: "application/json" })
pm.request.addHeader({ key: "X-New", value: "n" })
pm.request.removeHeader("X-Old")
pm.request.url.query.upsert({ key: "a", value: "2" })
pm.request.addQueryParams([{ key: "b", value: "3" }])
pm.request.body.update('{"x":1}')
pm.request.authorizeUsing("basic", { username: "u", password: "p" })`))
	if got := logTexts(run); got != "api.test /items a=1 /items?a=1 api.test https top t1" {
		t.Errorf("logs = %s", got)
	}
	sent := run.Request
	if sent.NormalizedMethod() != "PUT" || sent.URL != "https://api.test/items?a=2&b=3#top" || sent.Body.Raw.Content != `{"x":1}` || sent.Body.Raw.Language != rawlanguage.Json || sent.ID != "01R" {
		t.Errorf("sent = %+v", sent)
	}
	if !reflect.DeepEqual(sent.Headers, []entity.KeyValue{entity.NewKeyValue("accept", "application/json"), entity.NewKeyValue("X-New", "n")}) {
		t.Errorf("headers = %+v", sent.Headers)
	}
	if sent.Auth.Type != authtype.Basic || sent.Auth.Basic != (entity.AuthBasic{Username: "u", Password: "p"}) {
		t.Errorf("auth = %+v", sent.Auth)
	}
}

func TestReplacedPmRequestKeepsTheWrittenRequest(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test"
	for _, text := range []string{"pm.request.headers = null", "Object.defineProperty(pm.request, 'url', { value: { toString () { throw new Error('x') } } })"} {
		run := runIn(t, script(""), contextOf(written, nil), nil, text)
		if run.Request.URL != "https://api.test" || run.Request.ID != "01R" {
			t.Errorf("%s: request = %+v", text, run.Request)
		}
	}
}

func TestSkipRequestStopsThePreRequestScript(t *testing.T) {
	pre := runIn(t, script(""), entity.ScriptContext{}, nil, "console.log('before'); pm.execution.skipRequest(); console.log('after')")
	if pre.Status != scriptrunstatus.Skipped || logTexts(pre) != "before" || pre.Error != nil {
		t.Errorf("pre = %+v", pre)
	}
	inTimer := runIn(t, script(""), entity.ScriptContext{}, nil, "setTimeout(() => { pm.execution.skipRequest(); console.log('after') }, 1)")
	if inTimer.Status != scriptrunstatus.Skipped || logTexts(inTimer) != "" {
		t.Errorf("in timer = %+v", inTimer)
	}
	post := runIn(t, postScript(), entity.ScriptContext{}, nil, "console.log(typeof pm.execution.skipRequest); pm.execution.skipRequest()")
	if post.Status != scriptrunstatus.Error || post.Error.Message != "TypeError: pm.execution.skipRequest is not a function" || logTexts(post) != "function" {
		t.Errorf("post = %+v", post)
	}
}

func TestPmInfoAndLocation(t *testing.T) {
	scope := entity.ScriptContext{Info: entity.NewScriptInfo("01R", "Create", "Billing", []string{"Billing", "Create"})}
	run := ok(t, runIn(t, postScript(), scope, nil, `
console.log(pm.info.eventName, pm.info.iteration, pm.info.iterationCount, pm.info.requestName, pm.info.requestId, pm.execution.location.join("/"), pm.execution.location.current, pm.execution.setNextRequest("x"), iteration)`))
	if got := logTexts(run); got != "test 0 1 Create 01R Billing/Create Billing undefined 0" {
		t.Errorf("logs = %s", got)
	}
	bare := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, "console.log(pm.info.iteration, pm.info.requestName, pm.execution.location.length, pm.execution.location.current)"))
	if got := logTexts(bare); got != "undefined undefined 0 undefined" {
		t.Errorf("bare = %s", got)
	}
}

func TestPmVaultAndCookies(t *testing.T) {
	scope := entity.ScriptContext{
		Vault:           map[string]string{"key": "k1", "old": "o"},
		Cookies:         []entity.ScriptCookie{{Name: "session", Value: "s%201", Domain: "a.test", Path: "/", HttpOnly: true, HostOnly: true}},
		ResponseCookies: []entity.ScriptCookie{{Name: "session", Value: "s%201"}},
	}
	response := entity.NewResponse(200, "OK", "HTTP/1.1", nil, nil, 0, entity.Timings{})
	scope.Response = &response
	run := ok(t, runIn(t, postScript(), scope, nil, `
console.log(await pm.vault.get("key"), await pm.vault.get("nope"))
await pm.vault.set("new", 5)
await pm.vault.unset("old")
await pm.vault.set("cleared", null)
console.log(pm.cookies.get("session"), pm.cookies.one("SESSION").httpOnly, pm.cookies.one("session").toString(), pm.response.cookies.count())
try { await pm.cookies.jar().get("http://x.test", "a") } catch (e) { console.log(e.message) }`))
	if got := logTexts(run); got != "k1 null|s 1 true session=s%201; Path=/; HttpOnly 1|pm.cookies.jar is not available in this run" {
		t.Errorf("logs = %s", got)
	}
	if !reflect.DeepEqual(run.Vault, map[string]string{"key": "k1", "new": "5"}) {
		t.Errorf("vault = %v", run.Vault)
	}
}

func TestUnsupportedApis(t *testing.T) {
	run := runIn(t, script(""), entity.ScriptContext{}, nil, "console.log(typeof pm.datasets, typeof pm.require, typeof pm.mock); pm.datasets('x')")
	if run.Error == nil || run.Error.Kind != scripterrorkind.UnsupportedApi || run.Error.Message != "Error: pm.datasets is not supported in Xenon" || run.Error.Line != 1 {
		t.Errorf("datasets = %+v", run.Error)
	}
	if got := logTexts(run); got != "function undefined undefined" {
		t.Errorf("logs = %s", got)
	}
	required := runIn(t, script(""), entity.ScriptContext{}, nil, "\npm.require('x')")
	if required.Error == nil || required.Error.Kind != scripterrorkind.Runtime || required.Error.Message != "TypeError: pm.require is not a function" {
		t.Errorf("require = %+v", required.Error)
	}
	caught := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, "try { new SharedArrayBuffer(1) } catch (e) { console.log(e.message) }; try { Atomics.load() } catch (e) { console.log(e.message) }; console.log(Object.prototype.toString.call(new SharedArrayBuffer()))"))
	if got := logTexts(caught); got != "SharedArrayBuffer is not supported in Xenon|Atomics is not supported in Xenon|[object SharedArrayBuffer]" {
		t.Errorf("caught = %s", got)
	}
}

func TestPmTestsAndLegacyTests(t *testing.T) {
	response := entity.NewResponse(404, "Not Found", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}, []byte(`{"id":1}`), 8, entity.Timings{Total: 3 * time.Millisecond})
	run := ok(t, runIn(t, postScript(), entity.ScriptContext{Response: &response}, nil, `
pm.test("status", () => pm.response.to.have.status(200))
pm.test("json", () => pm.expect(pm.response.json().id).to.eql(1))
pm.test.skip("later")
pm.test("async", (done) => setTimeout(() => done(new Error("broke")), 1))
pm.test("no function")
tests["legacy"] = responseCode.code === 404
tests["legacy fails"] = 0
await null`))
	got := make([]string, 0, len(run.Tests))
	for _, test := range run.Tests {
		got = append(got, test.Name+"="+test.Status.String()+":"+test.ErrorMessage)
		if test.Script.Phase != scriptphase.PostResponse || test.Script.OwnerName != "Login" {
			t.Errorf("test %s has script %+v", test.Name, test.Script)
		}
	}
	want := []string{
		"status=failed:expected response to have status code 200 but got 404", "json=passed:", "later=skipped:", "no function=passed:",
		"legacy=passed:", "legacy fails=failed:expected 0 to be truthy", "async=failed:broke",
	}
	if !reflect.DeepEqual(got, want) || run.TestCount != 7 {
		t.Errorf("tests = %v", got)
	}
	if run.Tests[6].Async != true || run.Tests[4].Index != 5 || run.Tests[6].Status != teststatus.Failed {
		t.Errorf("tests = %+v", run.Tests)
	}
	nonLegacy := ok(t, runIn(t, postScript(), entity.ScriptContext{Response: &response}, nil, "'use sandbox2';\nconsole.log(typeof tests, typeof _, typeof postman, typeof JSON.parse)"))
	if got := logTexts(nonLegacy); got != "undefined undefined undefined function" {
		t.Errorf("non legacy = %s", got)
	}
}

func TestPmSendRequest(t *testing.T) {
	requests := &fakeRequests{reply: func(request entity.Request) entity.NestedRequest {
		if strings.Contains(request.URL, "fail") {
			return entity.NewRejectedNestedRequest("The request uses secrets: k")
		}
		if strings.Contains(request.URL, "down") {
			return entity.NewSentNestedRequest(entity.EffectiveRequest{}, entity.NewFailedResponse(entity.NewSendError(senderrorkind.ConnectionRefused, "x"), entity.Timings{}))
		}
		response := entity.NewResponse(200, "OK", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}, []byte(`{"token":"t"}`), 13, entity.Timings{Total: 5 * time.Millisecond})
		return entity.NewSentNestedRequest(entity.EffectiveRequest{URL: request.URL}, response)
	}}
	run := ok(t, runIn(t, script(""), contextOf(entity.NewRequest("", ""), map[string]any{"before": "1"}), requests, `
pm.variables.set("path", "token")
pm.environment.set("tenant", "globex")
const res = await pm.sendRequest({ url: "https://api.test/{{path}}", method: "post", header: { "X-A": "a" }, body: { mode: "raw", raw: "grant" } })
console.log(res.code, res.status, res.reason(), res.headers.get("content-type"), res.json().token, res.text(), res.responseTime, res.responseSize)
const chained = pm.sendRequest("https://api.test/fail", (err, response) => console.log("callback", err.message, response))
console.log(chained === pm)
try { await pm.sendRequest("https://api.test/down") } catch (err) { console.log("rejected", err.message !== "") }
try { await pm.sendRequest() } catch (err) { console.log(err.message) }
await new Promise((resolve) => setTimeout(resolve, 20))
pm.sendRequest("https://api.test/second", (err, response) => console.log("second", err, response.code))`))
	got := logTexts(run)
	for _, part := range []string{`200 OK OK application/json t {"token":"t"} 5 13`, "true", "callback The request uses secrets: k undefined", "rejected true", "sendrequest: nothing to request", "second null 200"} {
		if !strings.Contains(got, part) {
			t.Errorf("logs %q miss %q", got, part)
		}
	}
	if len(run.Nested) != 4 || len(requests.received) != 4 {
		t.Fatalf("nested = %+v", run.Nested)
	}
	first := requests.received[0]
	if first.Method != "POST" || first.URL != "https://api.test/{{path}}" || !reflect.DeepEqual(first.Headers, []entity.KeyValue{entity.NewKeyValue("X-A", "a")}) || first.Body.Raw.Content != "grant" || first.Body.Raw.Language != rawlanguage.Text {
		t.Errorf("first request = %+v", first)
	}
	if requests.scopes[0].Local.ValueMap()["path"] != "token" || requests.scopes[0].Environment.ValueMap()["tenant"] != "globex" || requests.scopes[0].Environment.Name != "dev" {
		t.Errorf("scopes at send time = %+v", requests.scopes[0])
	}
}

func TestPmSendRequestFailures(t *testing.T) {
	slow := &fakeRequests{reply: func(entity.Request) entity.NestedRequest {
		time.Sleep(300 * time.Millisecond)
		return entity.NewRejectedNestedRequest("late")
	}}
	start := time.Now()
	timedOut := testRunner(100*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script("pm.sendRequest('https://slow.test', () => console.log('never'))"), entity.ScriptContext{}, slow, nil, entity.DefaultAppSettings())
	if timedOut.Error == nil || timedOut.Error.Kind != scripterrorkind.Timeout || len(timedOut.Logs) != 0 || time.Since(start) > 2*time.Second {
		t.Errorf("timed out = %+v", timedOut)
	}
	time.Sleep(400 * time.Millisecond)
	echo := &fakeRequests{reply: func(entity.Request) entity.NestedRequest { return entity.NewRejectedNestedRequest("no") }}
	throwing := runIn(t, script(""), entity.ScriptContext{}, echo, "pm.sendRequest('https://x.test', () => { throw new Error('in callback') })")
	if throwing.Error == nil || throwing.Error.Message != "Unhandled promise rejection: Error: in callback" {
		t.Errorf("throwing = %+v", throwing.Error)
	}
	unavailable := runIn(t, script(""), entity.ScriptContext{}, nil, "pm.sendRequest('https://x.test')")
	if unavailable.Error == nil || unavailable.Error.Message != "Unhandled promise rejection: Error: pm.sendRequest is not available in this run" {
		t.Errorf("unavailable = %+v", unavailable.Error)
	}
}

func TestPmRunRequest(t *testing.T) {
	requests := &fakeRequests{run: func(requestId string, scopes entity.ScriptScopes) entity.NestedRequest {
		switch requestId {
		case "01OK":
			response := entity.NewResponse(201, "Created", "HTTP/1.1", nil, []byte(`{"ok":true}`), 11, entity.Timings{})
			execution := entity.NewSendExecution("child")
			execution.Start(entity.NewRequest(requestId, "Child"), scopes, nil)
			execution.Scopes.Environment = execution.Scopes.Environment.WithValue("token", "t-1")
			execution.Scopes.Local = execution.Scopes.Local.WithValue("child_only", "x")
			execution.RecordResponse(entity.EffectiveRequest{}, response)
			return entity.NewRunNestedRequest(requestId, execution)
		case "01SKIP":
			execution := entity.NewSendExecution("child")
			execution.Start(entity.NewRequest(requestId, "Child"), scopes, nil)
			execution.AddRun(entity.NewScriptRun(script(""), time.Time{}).WithEffects(entity.ScriptEffects{Scopes: scopes}).SkippedRequest())
			return entity.NewRunNestedRequest(requestId, execution)
		default:
			return entity.NewMissingRunNestedRequest(requestId, "runRequest: no saved request with id "+requestId)
		}
	}}
	run := ok(t, runIn(t, script(""), contextOf(entity.NewRequest("", ""), nil), requests, `
const res = await pm.execution.runRequest("01OK", { variables: { client: "c" } })
console.log(res.code, res.json().ok, pm.environment.get("token"), pm.variables.get("child_only"), pm.variables.get("client"))
console.log(await pm.execution.runRequest("01SKIP"))
try { await pm.execution.runRequest("01NO") } catch (e) { console.log(e.message) }
try { await pm.execution.runRequest() } catch (e) { console.log(e.message) }`))
	if got := logTexts(run); got != "201 true t-1 undefined undefined|null|runRequest: no saved request with id 01NO|runRequest: collection request id not provided" {
		t.Errorf("logs = %s", got)
	}
	if requests.scopes[0].Local.ValueMap()["client"] != "c" || len(run.Nested) != 3 || !run.Nested[0].IsRunRequest() {
		t.Errorf("scopes = %+v nested = %+v", requests.scopes[0], run.Nested)
	}
	if run.Scopes.Environment.ValueMap()["token"] != "t-1" {
		t.Errorf("merged scopes = %+v", run.Scopes.Environment)
	}
	unavailable := runIn(t, script(""), entity.ScriptContext{}, nil, "await pm.execution.runRequest('01X')")
	if unavailable.Error == nil || unavailable.Error.Message != "Error: pm.execution.runRequest is not available in this run" {
		t.Errorf("unavailable = %+v", unavailable.Error)
	}
}

func TestFailedScriptsKeepTheWrittenRequestAndTheirEffects(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test"
	run := runIn(t, script(""), contextOf(written, nil), nil, "pm.variables.set('a', 1); console.log('x'); pm.request.url = 'https://changed.test'; throw new Error('boom')")
	if !run.HasFailed() || run.Request.URL != "https://api.test" || run.Scopes.Local.ValueMap()["a"] != 1.0 || logTexts(run) != "x" {
		t.Errorf("run = %+v", run)
	}
	syntax := runIn(t, script(""), contextOf(written, map[string]any{"k": "v"}), nil, "const = 1")
	if syntax.Error.Kind != scripterrorkind.Syntax || syntax.Request.URL != "https://api.test" || syntax.Scopes.Local.ValueMap()["k"] != "v" || syntax.Logs == nil {
		t.Errorf("syntax = %+v", syntax)
	}
	timedOut := testRunner(100*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script("pm.variables.set('t', 1); console.log('y'); while (true) {}"), contextOf(written, nil), nil, nil, entity.DefaultAppSettings())
	if timedOut.Error.Kind != scripterrorkind.Timeout || timedOut.Scopes.Local.ValueMap()["t"] != 1.0 || logTexts(timedOut) != "y" || timedOut.Request.URL != "https://api.test" {
		t.Errorf("timed out = %+v", timedOut)
	}
}

func TestGlobalsBackedByGo(t *testing.T) {
	run := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `
console.log(btoa("ÿ"), atob(" aGk= "), atob("YQ"), new TextDecoder().decode(new TextEncoder().encode("é€")), new TextDecoder("utf8", { ignoreBOM: true }).decode(new Uint8Array([0xEF, 0xBB, 0xBF, 0x61])).length)
for (const value of ["%%", "a", "YQ=a"]) { try { atob(value) } catch (e) { console.log(e.name, e.code) } }
try { btoa("€") } catch (e) { console.log(e.name) }
try { new TextDecoder("utf-8", { fatal: true }).decode(new Uint8Array([0xff])) } catch (e) { console.log(e.name) }
console.log(new TextDecoder().decode(new Uint8Array([0xff])), new TextDecoder().decode(), new TextEncoder().encodeInto("ab", new Uint8Array(1)).written)
const digest = await crypto.subtle.digest("SHA-1", new Uint8Array([97]))
console.log(Buffer.from(digest).toString("hex"), crypto.getRandomValues(new Uint32Array(2)).length, crypto.randomUUID().length)
for (const bad of [new Float64Array(1), new Uint8Array(65537)]) { try { crypto.getRandomValues(bad) } catch (e) { console.log(e.name) } }
const key = await crypto.subtle.importKey("raw", new TextEncoder().encode("k"), { name: "HMAC", hash: "SHA-512" }, true, ["sign"])
const signature = await crypto.subtle.sign({ name: "HMAC" }, key, new TextEncoder().encode("m"))
console.log(new Uint8Array(signature).length, await crypto.subtle.verify("HMAC", key, new Uint8Array(64), new TextEncoder().encode("m")), new Uint8Array(await crypto.subtle.exportKey("raw", key))[0], key.algorithm.hash.name)
try { await crypto.subtle.digest("MD5", new Uint8Array(1)) } catch (e) { console.log(e.name) }
try { await crypto.subtle.sign("HMAC", {}, new Uint8Array(1)) } catch (e) { console.log(e.name) }
try { await crypto.subtle.importKey("jwk", {}, { name: "HMAC", hash: "SHA-1" }, false, []) } catch (e) { console.log(e.name) }
console.log(new URL("HTTP://Bücher.example:80/a/../b/./c?q=a b#f g").href, new URL("?x=1", "https://h.test/p").href, new URL("../up", "https://h.test/a/b/c").href, URL.canParse("nope"))`))
	want := "/w== hi a é€ 2|InvalidCharacterError 5|InvalidCharacterError 5|InvalidCharacterError 5|InvalidCharacterError|TypeError|�  1|86f7e437faa5a7fce15d1ddcb9eaeaea377667b8 2 36|TypeMismatchError|QuotaExceededError|64 false 107 SHA-512|NotSupportedError|TypeError|NotSupportedError|http://xn--bcher-kva.example/b/c?q=a%20b#f%20g https://h.test/p?x=1 https://h.test/a/up false"
	if got := logTexts(run); got != want {
		t.Errorf("logs =\n%s\nwant\n%s", got, want)
	}
}

func TestIntlFallsBackToEnUsWithANotice(t *testing.T) {
	run := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `
console.log((1234.5).toLocaleString("de-DE"), new Date(0).toLocaleDateString("en-US", { timeZone: "UTC" }), new Date(NaN).toLocaleString(), new Date(NaN).toLocaleDateString(), new Date(NaN).toLocaleTimeString(), new Intl.NumberFormat("fr").format(2), Intl.PluralRules !== undefined)
Intl = 5
console.log(Intl)`))
	if got := logTexts(run); got != "1,234.5 1/1/1970 Invalid Date Invalid Date Invalid Date 2 true|5" {
		t.Errorf("logs = %s", got)
	}
	if len(run.Notices) != 1 || run.Notices[0].Level != consolelevel.Warn || run.Notices[0].Text() != `Intl: the locale "de-DE" is not available in Xenon; en-US is used instead` {
		t.Errorf("notices = %+v", run.Notices)
	}
}

func TestSugarLoadsOnFirstUse(t *testing.T) {
	run := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `
console.log(Object.getOwnPropertyDescriptor(String.prototype, "has").get !== undefined)
console.log("abc".has("b"), [1, 2, 3].sum(), typeof Object.extended, Object.getOwnPropertyDescriptor(String.prototype, "has").get === undefined)
Array.prototype.first = 1
console.log([].first)`))
	if got := logTexts(run); got != "true|true 6 function true|1" {
		t.Errorf("logs = %s", got)
	}
	assigned := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, "String.prototype.has = () => 'mine'; console.log('x'.has())"))
	if logTexts(assigned) != "mine" {
		t.Errorf("assigned = %s", logTexts(assigned))
	}
}

func TestV8Messages(t *testing.T) {
	script := "pm.require('x')\nfoo.bar.baz()\nfor (const item of list) {}"
	cases := map[string]string{
		"Cannot read property 'x' of undefined":          "Cannot read properties of undefined (reading 'x')",
		"Cannot read property 'x' of undefined or null":  "Cannot read properties of undefined (reading 'x')",
		"TypeError: Object has no member 'require'":      "TypeError: pm.require is not a function",
		"Object has no member 'baz'":                     "foo.bar.baz is not a function",
		"Object has no member 'other'":                   "other is not a function",
		"object is not iterable":                         "list is not iterable",
		"plain message":                                  "plain message",
		"x Cannot read property 'a' of undefined y":      "x Cannot read properties of undefined (reading 'a') y",
		"Object has no member 'require' and has no more": "pm.require is not a function and has no more",
	}
	for input, want := range cases {
		if got := v8Message(input, script); got != want {
			t.Errorf("v8Message(%q) = %q, want %q", input, got, want)
		}
	}
	if got := calleeOf("a.run(); b.run()", "run"); got != "run" {
		t.Errorf("ambiguous callee = %q", got)
	}
	if got := iterableOf("for (const a of x) {} for (const b of y) {}"); got != "object" {
		t.Errorf("ambiguous iterable = %q", got)
	}
	if got := iterableOf("[...value]"); got != "object" {
		t.Errorf("no for-of = %q", got)
	}
}

func TestLocalZoneName(t *testing.T) {
	t.Setenv("TZ", "America/Sao_Paulo")
	if got := localZoneName(); got != "America/Sao_Paulo" {
		t.Errorf("zone = %q", got)
	}
	t.Setenv("TZ", "Not/A_Zone")
	if got := localZoneName(); got == "Not/A_Zone" || got == "" {
		t.Errorf("an unknown TZ falls back: %q", got)
	}
}
