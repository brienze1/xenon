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
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

type fakeRequests struct {
	reply    func(entity.Request) entity.NestedRequest
	received []entity.Request
	locals   []map[string]any
	mutex    sync.Mutex
}

func (f *fakeRequests) Send(_ context.Context, request entity.Request, localVariables map[string]any) entity.NestedRequest {
	f.mutex.Lock()
	f.received = append(f.received, request)
	f.locals = append(f.locals, localVariables)
	f.mutex.Unlock()
	return f.reply(request)
}

func scopeOf(request entity.Request, local map[string]any) entity.ScriptContext {
	variables := entity.NewVariableContext(
		entity.NewVariableLayer(variablescope.Environment, []entity.VariableRow{entity.NewVariableRow("tenant", "acme"), {Key: "vaulted", Secret: true, Enabled: true}}, nil, nil, true),
	)
	return entity.ScriptContext{Request: request, LocalVariables: local, Variables: variables}
}

func runIn(t *testing.T, phase entity.ScriptRef, scope entity.ScriptContext, requests usecase.SendScriptRequest, text string) entity.ScriptRun {
	t.Helper()
	phase.Text = text
	return testRunner(5*time.Second, memoryGrowthLimit).Run(context.Background(), phase, scope, requests, entity.DefaultAppSettings())
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

func TestPmVariables(t *testing.T) {
	run := ok(t, runIn(t, script(""), scopeOf(entity.NewRequest("", ""), map[string]any{"kept": "yes", "gone": 1.0}), nil, `
pm.variables.set("n", 2)
pm.variables.set("obj", { a: [1] })
pm.variables.set("nothing", undefined)
pm.variables.unset("gone")
console.log(pm.variables.get("kept"), pm.variables.get("n") + 1, pm.variables.get("obj").a[0], pm.variables.get("nothing"), pm.variables.get("gone"))
console.log(pm.variables.get("tenant"), pm.variables.has("tenant"), pm.variables.has("vaulted"), pm.variables.get("vaulted"), pm.variables.has("missing"))
console.log(pm.variables.replaceIn("{{tenant}}/{{n}}/{{kept}}/{{unknown}}"))
console.log(JSON.stringify(pm.variables.toObject()))`))
	want := `yes 3 1 null undefined|acme true false undefined false|acme/2/yes/{{unknown}}|{"kept":"yes","n":2,"nothing":null,"obj":{"a":[1]},"tenant":"acme"}`
	if got := logTexts(run); got != want {
		t.Errorf("logs = %s", got)
	}
	if !reflect.DeepEqual(run.LocalVariables, map[string]any{"kept": "yes", "n": 2.0, "obj": map[string]any{"a": []any{1.0}}, "nothing": nil}) {
		t.Errorf("local = %#v", run.LocalVariables)
	}
	cleared := ok(t, runIn(t, script(""), scopeOf(entity.NewRequest("", ""), map[string]any{"kept": "yes"}), nil, `pm.variables.clear(); console.log(pm.variables.has("kept"))`))
	if logTexts(cleared) != "false" || len(cleared.LocalVariables) != 0 {
		t.Errorf("clear = %s %v", logTexts(cleared), cleared.LocalVariables)
	}
	noContext := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `pm.variables.set("a", "b")`))
	if noContext.LocalVariables["a"] != "b" {
		t.Errorf("a missing local scope starts empty: %v", noContext.LocalVariables)
	}
}

func TestPmRequestChangesComeBackToTheSend(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test/items?a=1#top"
	written.Headers = []entity.KeyValue{entity.NewKeyValue("X-Old", "1"), entity.NewKeyValue("Accept", "text/plain")}
	written.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Json, Content: "{}"}}
	run := ok(t, runIn(t, script(""), scopeOf(written, nil), nil, `
const url = pm.request.url
console.log(url.getHost(), url.getPath(), url.getQueryString(), url.getPathWithQuery(), url.getRemote(), url.protocol, url.hash)
pm.request.method = "put"
pm.request.headers.upsert({ key: "accept", value: "application/json" })
pm.request.addHeader({ key: "X-New", value: "n" })
pm.request.upsertHeader({ key: "X-Up", value: "u" })
pm.request.removeHeader("x-old")
pm.request.url.query.upsert({ key: "a", value: "2" })
pm.request.removeQueryParams("missing")
pm.request.addQueryParams([{ key: "b", value: "3" }])
console.log(pm.request.headers.count(), pm.request.headers.has("X-NEW"), pm.request.headers.get("ACCEPT"), JSON.stringify(pm.request.getHeaders()))
pm.request.body.update('{"x":1}')
console.log(pm.request.body.toString(), pm.request.body.isEmpty(), pm.request.url.toString())`))
	if got := logTexts(run); got != `api.test /items a=1 /items?a=1 api.test https top|3 true application/json {"accept":"application/json","X-New":"n","X-Up":"u"}|{"x":1} false https://api.test/items?a=2&b=3#top` {
		t.Errorf("logs = %s", got)
	}
	sent := run.Request
	if sent.NormalizedMethod() != "PUT" || sent.URL != "https://api.test/items?a=2&b=3#top" || sent.Body.Raw.Content != `{"x":1}` || sent.Body.Raw.Language != rawlanguage.Json || sent.ID != "01R" {
		t.Errorf("sent = %+v", sent)
	}
	if !reflect.DeepEqual(sent.Headers, []entity.KeyValue{entity.NewKeyValue("accept", "application/json"), entity.NewKeyValue("X-New", "n"), entity.NewKeyValue("X-Up", "u")}) {
		t.Errorf("headers = %+v", sent.Headers)
	}
}

func TestPropertyListAndUrlShims(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "http://localhost:8080/a/b?flag&x=1&x=2"
	written.Headers = []entity.KeyValue{{Key: "X-Off", Value: "0"}}
	run := ok(t, runIn(t, script(""), scopeOf(written, nil), nil, `
const q = pm.request.url.query
console.log(q.count(), q.get("x"), q.has("x", "1"), q.has("x", "9"), q.has("flag"), q.idx(0).toString(), q.indexOf(q.idx(1)), q.all().length)
console.log(pm.request.url.getRemote(), pm.request.url.port, q.toString(), JSON.stringify(q.toObject()))
q.prepend("first=0")
q.insert({ key: "mid", value: "m" }, q.idx(2))
q.insert({ key: "last", value: "l" }, null)
q.append({ key: "app", value: "a" })
q.add(null)
console.log(q.map((p) => p.key).join(), q.filter((p) => p.key === "x").length, q.find((p) => p.key === "mid").value, q.reduce((total, p) => total + p.key.length, 0))
let keys = []
q.each((p) => keys.push(p.key))
q.remove((p) => p.key === "x")
q.remove(q.idx(0))
console.log(keys.length, q.toString(), q.upsert(null))
q.repopulate([{ key: "only", value: "1", disabled: true }])
console.log(q.toString(), JSON.stringify(q.toObject(true)), JSON.stringify(q.toJSON()), q.idx(0).valueOf())
const h = pm.request.headers
h.add("X-Line: v1")
h.add("Bare")
console.log(h.toString().split("\n").length, h.get("x-line"), h.one("bare").value, JSON.stringify(h.toObject()))
pm.request.removeQueryParams({ key: "only" })
pm.request.url.update("https://other.test/z")
console.log(pm.request.url.toString(), pm.request.url.getRaw(), JSON.stringify(pm.request.url.toJSON()), pm.request.url.getPath())
h.clear()
console.log(h.count(), JSON.stringify(pm.request.toJSON().header))`))
	want := `3 2 true false true flag 1 3|localhost:8080 8080 flag&x=1&x=2 {"flag":null,"x":"2"}|first,flag,mid,x,x,last,app 2 m 21|7 flag&mid=m&last=l&app=a null|` + `` + ` {} [{"key":"only","value":"1","disabled":true}] 1|2 v1 null {"X-Off":"0","X-Line":"v1","Bare":null}|https://other.test/z https://other.test/z "https://other.test/z" /z|0 []`
	if got := logTexts(run); got != want {
		t.Errorf("logs =\n%s\nwant\n%s", got, want)
	}
}

func TestRequestBodyModes(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.Body = entity.Body{Mode: bodymode.UrlEncoded, UrlEncoded: []entity.KeyValue{entity.NewKeyValue("k", "v")}}
	run := ok(t, runIn(t, script(""), scopeOf(written, nil), nil, `
const body = pm.request.body
console.log(body.mode, body.isEmpty(), body.toString(), JSON.stringify(body.toJSON()))
body.update({ mode: "formdata", formdata: [{ key: "f", value: "1" }] })
console.log(body.isEmpty(), body.toString() === "", JSON.stringify(body.toJSON()))
for (const mode of ["raw", "urlencoded", "formdata", "graphql", "file", undefined]) {
  body.update({ mode })
  console.log(mode, body.isEmpty())
}
body.update({ mode: "graphql", graphql: { query: "q" } })
console.log(body.isEmpty(), body.toString() === "")
body.update({ mode: "file", file: { src: "x" } })
console.log(body.isEmpty(), JSON.stringify(body.toJSON()))
body.update({ mode: "raw" })
console.log(body.toString() === "", JSON.stringify(body.toJSON()))
body.update({ mode: "urlencoded", urlencoded: [{ key: "a", value: "1" }] })`))
	want := `urlencoded false k=v {"mode":"urlencoded","urlencoded":[{"key":"k","value":"v"}]}|false true {"mode":"formdata","formdata":[{"key":"f","value":"1"}]}|raw true|urlencoded true|formdata true|graphql true|file true|undefined true|false true|false {"mode":"file"}|true {"mode":"raw","raw":""}`
	if got := logTexts(run); got != want {
		t.Errorf("logs =\n%s\nwant\n%s", got, want)
	}
	if run.Request.Body.Mode != bodymode.UrlEncoded || !reflect.DeepEqual(run.Request.Body.UrlEncoded, []entity.KeyValue{entity.NewKeyValue("a", "1")}) {
		t.Errorf("body = %+v", run.Request.Body)
	}
}

func TestReplacedPmRequestKeepsTheWrittenRequest(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test"
	for _, text := range []string{"pm.request = null", "pm.request = { toJSON: () => 5 }"} {
		run := ok(t, runIn(t, script(""), scopeOf(written, nil), nil, text))
		if run.Request.URL != "https://api.test" || run.Request.ID != "01R" {
			t.Errorf("%s: request = %+v", text, run.Request)
		}
	}
}

func TestSkipRequestOnlyInPreRequestScripts(t *testing.T) {
	pre := runIn(t, script(""), entity.ScriptContext{}, nil, "pm.execution.skipRequest(); console.log('after')")
	if pre.Status != scriptrunstatus.Skipped || logTexts(pre) != "after" {
		t.Errorf("pre = %+v", pre)
	}
	post := script("")
	post.Phase = scriptphase.PostResponse
	if run := runIn(t, post, entity.ScriptContext{}, nil, "pm.execution.skipRequest()"); run.Status != scriptrunstatus.Ok {
		t.Errorf("post = %+v", run)
	}
	failed := runIn(t, script(""), entity.ScriptContext{}, nil, "pm.execution.skipRequest(); throw new Error('x')")
	if failed.Status != scriptrunstatus.Error {
		t.Errorf("an error wins over skip: %+v", failed)
	}
}

func TestConsoleArgumentsAreDescribedWhenLogged(t *testing.T) {
	run := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, `
const shared = { n: 1 }
const cycle = {}
cycle.self = cycle
console.log("text", 1, true, null, undefined, shared, [1, "a"], new Error("e"), function named() {}, cycle, { toJSON: () => undefined }, Symbol("s"))
shared.n = 2
console.warn()`))
	if len(run.Logs) != 2 || run.Logs[0].Level != consolelevel.Log || run.Logs[1].Level != consolelevel.Warn || len(run.Logs[1].Arguments) != 0 {
		t.Fatalf("logs = %+v", run.Logs)
	}
	want := []string{"text", "1", "true", "null", "undefined", `{"n":1}`, `[1,"a"]`, "Error: e", "function named() {}", "[object Object]", "[object Object]", "s"}
	if !reflect.DeepEqual(run.Logs[0].Arguments, want) {
		t.Errorf("arguments = %#v", run.Logs[0].Arguments)
	}
	cleared := ok(t, runIn(t, script(""), entity.ScriptContext{}, nil, "console.log('a'); console.clear(); console.info('b')"))
	if logTexts(cleared) != "b" {
		t.Errorf("clear = %s", logTexts(cleared))
	}
}

func TestFailedScriptsKeepTheWrittenRequestAndTheirEffects(t *testing.T) {
	written := entity.NewRequest("01R", "Create")
	written.URL = "https://api.test"
	run := runIn(t, script(""), scopeOf(written, nil), nil, "pm.variables.set('a', 1); console.log('x'); pm.request.url = 'https://changed.test'; throw new Error('boom')")
	if !run.HasFailed() || run.Request.URL != "https://api.test" || run.LocalVariables["a"] != 1.0 || logTexts(run) != "x" {
		t.Errorf("run = %+v", run)
	}
	syntax := runIn(t, script(""), scopeOf(written, map[string]any{"k": "v"}), nil, "const = 1")
	if syntax.Error.Kind != scripterrorkind.Syntax || syntax.Request.URL != "https://api.test" || syntax.LocalVariables["k"] != "v" || syntax.Logs == nil {
		t.Errorf("syntax = %+v", syntax)
	}
	timedOut := testRunner(100*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script("pm.variables.set('t', 1); console.log('y'); while (true) {}"), scopeOf(written, nil), nil, entity.DefaultAppSettings())
	if timedOut.Error.Kind != scripterrorkind.Timeout || timedOut.LocalVariables["t"] != 1.0 || logTexts(timedOut) != "y" || timedOut.Request.URL != "https://api.test" {
		t.Errorf("timed out = %+v", timedOut)
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
	run := ok(t, runIn(t, script(""), entity.ScriptContext{LocalVariables: map[string]any{"before": "1"}}, requests, `
pm.variables.set("path", "token")
const res = await pm.sendRequest({ url: "https://api.test/{{path}}", method: "post", header: { "X-A": "a" }, body: { mode: "raw", raw: "grant" } })
console.log(res.code, res.status, res.reason(), res.headers.get("content-type"), res.json().token, res.text(), res.responseTime, res.responseSize, JSON.stringify(res.toJSON()).length > 0)
pm.sendRequest("https://api.test/fail", (err, response) => console.log("callback", err.message, response))
try { await pm.sendRequest("https://api.test/down") } catch (err) { console.log("rejected", err.message !== "") }
await new Promise((resolve) => setTimeout(resolve, 20))
pm.sendRequest("https://api.test/second", (err, response) => console.log("second", err, response.code))`))
	if got := logTexts(run); !strings.HasPrefix(got, `200 OK OK application/json t {"token":"t"} 5 13 true|`) || !strings.Contains(got, "callback The request uses secrets: k undefined") || !strings.Contains(got, "rejected true") || !strings.HasSuffix(got, "second null 200") {
		t.Errorf("logs = %s", got)
	}
	if len(run.Nested) != 4 || len(requests.received) != 4 {
		t.Fatalf("nested = %+v", run.Nested)
	}
	first := requests.received[0]
	if first.Method != "POST" || first.URL != "https://api.test/{{path}}" || !reflect.DeepEqual(first.Headers, []entity.KeyValue{entity.NewKeyValue("X-A", "a")}) || first.Body.Raw.Content != "grant" || first.Body.Raw.Language != rawlanguage.Text {
		t.Errorf("first request = %+v", first)
	}
	if !reflect.DeepEqual(requests.locals[0], map[string]any{"before": "1", "path": "token"}) {
		t.Errorf("local variables at send time = %v", requests.locals[0])
	}
}

func TestPmSendRequestFailures(t *testing.T) {
	slow := &fakeRequests{reply: func(entity.Request) entity.NestedRequest {
		time.Sleep(300 * time.Millisecond)
		return entity.NewRejectedNestedRequest("late")
	}}
	start := time.Now()
	timedOut := testRunner(100*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script("pm.sendRequest('https://slow.test', () => console.log('never'))"), entity.ScriptContext{}, slow, entity.DefaultAppSettings())
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
	invalid := runIn(t, script(""), entity.ScriptContext{}, echo, "pm.request.toJSON = () => ({ method: 5 }); await pm.sendRequest(pm.request)")
	if invalid.Error == nil || invalid.Error.Message != "TypeError: pm.sendRequest: the request is not valid" {
		t.Errorf("invalid = %+v", invalid.Error)
	}
}
