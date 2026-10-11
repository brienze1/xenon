package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
)

func testRunner(timeout time.Duration, memoryGrowth uint64) *scriptRunnerWebservice {
	return &scriptRunnerWebservice{modules: sharedModuleRegistry(), limits: func(entity.AppSettings) limits {
		return limits{timeout: timeout, pollInterval: 10 * time.Millisecond, memoryGrowth: memoryGrowth}
	}}
}

func script(text string) entity.ScriptRef {
	return entity.ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Request, ItemID: "01R", OwnerName: "Login", Text: text}
}

func run(t *testing.T, text string) entity.ScriptRun {
	t.Helper()
	return testRunner(30*time.Second, memoryGrowthLimit).Run(context.Background(), script(text), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
}

func mustSucceed(t *testing.T, text string) {
	t.Helper()
	if result := run(t, text); result.HasFailed() {
		t.Fatalf("script failed: %+v", *result.Error)
	}
}

func mustFail(t *testing.T, text string, kind enums.ScriptErrorKind, message string, line, column int) entity.ScriptError {
	t.Helper()
	result := run(t, text)
	if !result.HasFailed() || result.Error == nil {
		t.Fatalf("script %q did not fail", text)
	}
	got := *result.Error
	if got.Kind != kind || got.Message != message || got.Line != line || got.Column != column {
		t.Fatalf("error = %+v; want kind %s, message %q at %d:%d", got, kind, message, line, column)
	}
	if got.Script != script(text) {
		t.Fatalf("error script = %+v", got.Script)
	}
	return got
}

func registryWith(extra map[string]string) *moduleRegistry {
	files := map[string]string{}
	for _, entry := range readManifest().Files {
		files[entry.Module] = entry.File
	}
	for module, file := range extra {
		files[module] = file
	}
	return &moduleRegistry{files: files}
}

func TestNewScriptRunnerWebserviceRunsScripts(t *testing.T) {
	result := NewScriptRunnerWebservice().Run(context.Background(), script("const a = 1 + 1; if (a !== 2) { throw new Error('math') }"), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
	if result.Status != scriptrunstatus.Ok || result.Error != nil || result.Script.Text == "" || result.Duration < 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestDefaultLimits(t *testing.T) {
	got := newLimits(entity.AppSettings{ScriptTimeoutMs: 1500})
	if got.timeout != 1500*time.Millisecond || got.memoryGrowth != 512<<20 || got.pollInterval != 50*time.Millisecond {
		t.Fatalf("limits = %+v", got)
	}
}

func TestPlainJavaScriptAndTopLevelAwait(t *testing.T) {
	mustSucceed(t, `
class Point { #x = 1; get x() { return this.#x } }
const total = [1, 2, 3].reduce((a, b) => a + b, 0);
const value = await new Promise((resolve) => setTimeout(() => resolve(total), 5));
if (new Point().x !== 1 || value !== 6 || (null ?? 'd') !== 'd') { throw new Error('language') }
return;`)
}

func TestEachRunHasItsOwnScope(t *testing.T) {
	mustSucceed(t, "globalThis.leaked = 1; var shared = 2;")
	mustSucceed(t, "if (typeof leaked !== 'undefined' || typeof shared !== 'undefined') { throw new Error('leak') }")
}

func TestSyntaxErrorsReportUserLineAndColumn(t *testing.T) {
	mustFail(t, "const a = 1;\nconst b = ;", scripterrorkind.Syntax, "SyntaxError: Unexpected token ';'", 2, 11)
	mustFail(t, "let = = 2", scripterrorkind.Syntax, "SyntaxError: Unexpected token '='", 1, 7)
}

func TestRuntimeErrorsReportUserLineAndColumn(t *testing.T) {
	mustFail(t, "const a = 1;\n  null.x;", scripterrorkind.Runtime, "TypeError: Cannot read properties of undefined (reading 'x')", 2, 8)
	mustFail(t, "throw new Error('boom')", scripterrorkind.Runtime, "Error: boom", 1, 7)
	mustFail(t, "await null;\nthrow new RangeError('late')", scripterrorkind.Runtime, "RangeError: late", 2, 7)
	mustFail(t, "throw 'plain'", scripterrorkind.Runtime, "plain", 0, 0)
	mustFail(t, "throw {message: 'object'}", scripterrorkind.Runtime, "[object Object]", 0, 0)
}

func TestErrorsInTimerCallbacksStopTheScript(t *testing.T) {
	mustFail(t, "setTimeout(() => {\n  undefinedFn();\n}, 1);\nsetTimeout(() => { throw new Error('second') }, 50);", scripterrorkind.Runtime, "ReferenceError: undefinedFn is not defined", 2, 14)
}

func TestTimers(t *testing.T) {
	mustSucceed(t, `
let order = [];
const cancelled = setTimeout(() => order.push('cancelled'), 1);
clearTimeout(cancelled);
clearTimeout(9999);
setImmediate((a, b) => order.push('immediate' + a + b), 1, 2);
setTimeout((x) => order.push('timeout' + x), 2, 'x');
let ticks = 0;
const interval = setInterval(() => { ticks++; if (ticks === 3) { clearInterval(interval) } }, 1);
await new Promise((resolve) => setTimeout(resolve, 40));
if (order.join() !== 'immediate12,timeoutx' || ticks !== 3) { throw new Error(order.join() + ticks) }`)
	mustFail(t, "setTimeout('code')", scripterrorkind.Runtime, `TypeError: The "callback" argument must be of type function`, 1, 11)
}

func TestScriptCompletesWhenTimersAreDone(t *testing.T) {
	start := time.Now()
	mustSucceed(t, "setTimeout(() => {}, 30); new Promise(() => {});")
	if elapsed := time.Since(start); elapsed < 30*time.Millisecond {
		t.Fatalf("did not wait for the pending timer: %v", elapsed)
	}
}

func TestTimeoutStopsBusyAndWaitingScripts(t *testing.T) {
	for _, text := range []string{
		"while (true) {}",
		"function tick() { setTimeout(tick, 1) } tick();",
		"setTimeout(() => { while (true) {} }, 1)",
		"setInterval(() => {}, 1)",
		"await new Promise((resolve) => setTimeout(resolve, 60000))",
	} {
		start := time.Now()
		result := testRunner(200*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script(text), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
		elapsed := time.Since(start)
		if result.Error == nil || result.Error.Kind != scripterrorkind.Timeout || result.Error.Message != "Script execution timed out" {
			t.Fatalf("%q: result = %+v", text, result)
		}
		if elapsed > 1500*time.Millisecond {
			t.Fatalf("%q stopped after %v", text, elapsed)
		}
	}
}

func TestStackOverflow(t *testing.T) {
	result := run(t, "function down(n) { return down(n + 1) + 1 }\ndown(0)")
	if result.Error == nil || result.Error.Kind != scripterrorkind.StackOverflow || result.Error.Message != "RangeError: Maximum call stack size exceeded" || result.Error.Line != 1 {
		t.Fatalf("result = %+v", result.Error)
	}
}

func TestMemoryGuard(t *testing.T) {
	result := testRunner(20*time.Second, 16<<20).Run(context.Background(), script("const keep = []; while (true) { keep.push(new Array(100000).fill('xxxxxxxx')) }"), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
	if result.Error == nil || result.Error.Kind != scripterrorkind.Memory || result.Error.Message != "Script stopped: it used too much memory" {
		t.Fatalf("result = %+v", result.Error)
	}
}

func TestUnhandledRejections(t *testing.T) {
	mustFail(t, "Promise.reject(new TypeError('lost'))", scripterrorkind.UnhandledRejection, "Unhandled promise rejection: TypeError: lost", 1, 16)
	mustFail(t, "Promise.reject('text')", scripterrorkind.UnhandledRejection, "Unhandled promise rejection: text", 0, 0)
	mustSucceed(t, "const p = Promise.reject(new Error('x')); p.catch(() => {});")
	mustSucceed(t, "const p = Promise.reject(new Error('x')); setTimeout(() => p.catch(() => {}), 1);")
	mustSucceed(t, "try { await Promise.reject(new Error('x')) } catch (e) {}")
}

func TestRequire(t *testing.T) {
	mustFail(t, "require('child_process')", scripterrorkind.Runtime, "Error: Cannot find module 'child_process'", 1, 8)
	mustFail(t, "require('liquid-json')", scripterrorkind.Runtime, "Error: Cannot find module 'liquid-json'", 1, 8)
	mustSucceed(t, "if (Object.keys(require('fs')).length !== 0 || typeof require('os') !== 'object') { throw new Error('fs and os are empty, as in postman-sandbox') }")
	mustSucceed(t, "if (require('atob')('aGk=') !== 'hi' || require('btoa')('hi') !== 'aGk=') { throw new Error('atob/btoa modules') }")
	mustSucceed(t, "const sdk = require('postman-collection'); if (new sdk.Url('https://a.io/x').getHost() !== 'a.io' || typeof sdk.VariableScope !== 'function') { throw new Error('postman-collection') }")
	mustFail(t, "require('lodash3')", scripterrorkind.Runtime, "Error: Cannot find module 'lodash3'", 1, 8)
	mustFail(t, "require('intl')", scripterrorkind.Runtime, "Error: Cannot find module 'intl'", 1, 8)
	mustSucceed(t, "if (require('lodash') !== require('lodash')) { throw new Error('not cached') }")
}

func TestEveryRequirableModuleLoads(t *testing.T) {
	checks := map[string]string{
		"ajv":                "new m({allErrors: true}).validate({type: 'object', required: ['a']}, {a: 1}) === true",
		"assert":             "typeof m.strictEqual === 'function' && (() => { try { m.strictEqual(1, 2); return false } catch (e) { return e.name === 'AssertionError' } })()",
		"backbone":           "typeof m.Model === 'function' && new m.Model({a: 1}).get('a') === 1",
		"buffer":             "m.Buffer.from('aGk=', 'base64').toString() === 'hi'",
		"chai":               "(() => { m.expect(1).to.equal(1); return true })()",
		"cheerio":            "m.load('<p class=\"x\">hey</p>')('.x').text() === 'hey'",
		"crypto-js":          "m.HmacSHA256('msg', 'key').toString() === '2d93cbc1be167bcb1637a4a23cbff01a7878f0c50ee833954ea5221bb1b8c628'",
		"csv-parse/lib/sync": "m('a,b\\n1,2', {columns: true})[0].b === '2'",
		"events":             "(() => { const e = new m(); let got = 0; e.on('x', (v) => got = v); e.emit('x', 3); return got === 3 })()",
		"lodash":             "m.chunk([1, 2, 3], 2).length === 2 && m.VERSION === '4.18.1'",
		"moment":             "m.utc('2024-01-31').add(1, 'month').format('YYYY-MM-DD') === '2024-02-29'",
		"path":               "m.join('/a', 'b', '../c') === '/a/c'",
		"punycode":           "m.toASCII('mañana.com') === 'xn--maana-pta.com'",
		"querystring":        "m.stringify({a: [1, 2], b: 'x y'}) === 'a=1&a=2&b=x%20y'",
		"stream":             "typeof m.Readable === 'function'",
		"string_decoder":     "new m.StringDecoder('utf8').write(require('buffer').Buffer.from([0xe2, 0x82, 0xac])) === '€'",
		"timers":             "typeof m.setTimeout === 'function'",
		"tv4":                "m.validate({a: 1}, {properties: {a: {type: 'number'}}}) === true",
		"url":                "m.parse('https://u:p@x.test:8080/p?q=1').port === '8080'",
		"util":               "m.format('%s=%d', 'a', 1) === 'a=1'",
		"uuid":               "/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(m.v4())",
		"xml2js":             "(() => { let out; m.parseString('<a><b>1</b></a>', (err, res) => { out = res }); return out.a.b[0] === '1' })()",
	}
	for _, entry := range readManifest().Files {
		if internalModules[entry.Module] {
			continue
		}
		check, found := checks[entry.Module]
		if !found {
			t.Fatalf("no load check for module %q", entry.Module)
		}
		t.Run(entry.Module, func(t *testing.T) {
			mustSucceed(t, "const m = require('"+entry.Module+"'); if (!("+check+")) { throw new Error('unexpected result') }")
		})
	}
}

func TestIntlAssetLoads(t *testing.T) {
	runner := testRunner(time.Minute, memoryGrowthLimit)
	runner.modules = registryWith(map[string]string{"intl-test": "intl.js"})
	result := runner.Run(context.Background(), script(`
const m = require('intl-test');
const n = new m.Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'}).format(1234.5);
const d = new m.Intl.DateTimeFormat('en-US', {timeZone: 'UTC', year: 'numeric', month: 'long', day: 'numeric'}).format(new Date(Date.UTC(2024, 1, 29)));
if (n !== '$1,234.50' || d !== 'February 29, 2024') { throw new Error(n + ' ' + d) }`), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
	if result.HasFailed() {
		t.Fatalf("intl: %+v", *result.Error)
	}
}

func TestLazyLibraryGlobals(t *testing.T) {
	mustSucceed(t, `
if (CryptoJS.MD5('a').toString() !== '0cc175b9c0f1b6a831c399e269772661') { throw new Error('CryptoJS') }
if (_.VERSION !== '3.10.2' || typeof _.pluck !== 'function') { throw new Error('lodash3') }
if (!tv4.validate(1, {type: 'number'})) { throw new Error('tv4') }
if (cheerio.load('<b>x</b>')('b').text() !== 'x') { throw new Error('cheerio') }
if (typeof Backbone.Collection !== 'function') { throw new Error('Backbone') }
if (Buffer.from('hi').toString('base64') !== 'aGk=') { throw new Error('Buffer') }
CryptoJS = 'mine';
if (CryptoJS !== 'mine') { throw new Error('overwrite') }`)
	mustSucceed(t, "tv4 = 1; if (tv4 !== 1) { throw new Error('assign before read') }")
}

func TestModuleErrorsSurfaceAsScriptErrors(t *testing.T) {
	registry := registryWith(map[string]string{"broken": "missing.js"})
	runner := &scriptRunnerWebservice{modules: registry, limits: newLimits}
	result := runner.Run(context.Background(), script("require('broken')"), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
	if result.Error == nil || !strings.HasPrefix(result.Error.Message, "Error: Cannot load module 'broken': ") {
		t.Fatalf("result = %+v", result.Error)
	}
}

func TestStopperForwardsToRuntimeAndLoop(t *testing.T) {
	result := testRunner(50*time.Millisecond, memoryGrowthLimit).Run(context.Background(), script("for (;;) {}"), entity.ScriptContext{}, nil, nil, entity.DefaultAppSettings())
	if result.Error == nil || result.Error.Kind != scripterrorkind.Timeout {
		t.Fatalf("result = %+v", result.Error)
	}
}
