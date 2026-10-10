package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

var sandboxAllowlist = []string{
	"pm",
	"console", "setTimeout", "setInterval", "setImmediate", "clearTimeout", "clearInterval", "clearImmediate",
	"atob", "btoa", "Buffer", "URL", "URLSearchParams", "TextEncoder", "TextDecoder", "structuredClone",
	"queueMicrotask", "crypto", "AbortController", "AbortSignal", "Event", "EventTarget", "DOMException",
	"Blob", "Intl", "require",
	"tests", "globals", "environment", "data", "request", "responseCookies", "responseHeaders",
	"responseTime", "responseCode", "responseBody", "iteration", "postman", "CryptoJS", "tv4", "xml2Json",
	"cheerio", "_", "Backbone",
	"SharedArrayBuffer", "Atomics",
}

var pendingGlobals = []string{}

func ecmaScriptGlobals(t *testing.T) []string {
	t.Helper()
	value, err := goja.New().RunString("JSON.stringify(Object.getOwnPropertyNames(globalThis))")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err = json.Unmarshal([]byte(value.String()), &names); err != nil {
		t.Fatal(err)
	}
	return names
}

func jsonList(t *testing.T, values []string) string {
	t.Helper()
	encoded, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestGlobalObjectMatchesTheSandboxAllowlist(t *testing.T) {
	installed := slices.DeleteFunc(slices.Clone(sandboxAllowlist), func(name string) bool { return slices.Contains(pendingGlobals, name) })
	requires := make([]string, 0)
	for _, entry := range readManifest().Files {
		if !internalModules[entry.Module] {
			requires = append(requires, "require('"+entry.Module+"');")
		}
	}
	text := strings.Join(requires, "\n") + `
const builtins = ` + jsonList(t, ecmaScriptGlobals(t)) + `;
const allowed = ` + jsonList(t, installed) + `;
[CryptoJS, tv4, cheerio, _, Backbone, Buffer].forEach((value) => { if (value === undefined) { throw new Error('lazy global missing') } });
const names = Object.getOwnPropertyNames(globalThis);
const extra = names.filter((name) => !builtins.includes(name) && !allowed.includes(name));
const missing = allowed.filter((name) => !names.includes(name));
if (extra.length || missing.length) { throw new Error('extra=' + extra.join(',') + ' missing=' + missing.join(',')) }`
	postResponse := script(text)
	postResponse.Phase = scriptphase.PostResponse
	response := entity.NewResponse(200, "OK", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}, []byte("{}"), 2, entity.Timings{})
	result := testRunner(time.Minute, memoryGrowthLimit).Run(context.Background(), postResponse, entity.ScriptContext{Response: &response}, nil, entity.DefaultAppSettings())
	if result.HasFailed() {
		t.Fatalf("global object differs from the allowlist: %s", result.Error.Message)
	}
	for _, name := range pendingGlobals {
		if !slices.Contains(sandboxAllowlist, name) {
			t.Errorf("pending global %q is not in the allowlist", name)
		}
	}
}

func TestGlobalsExposeNoGoObjects(t *testing.T) {
	mustSucceed(t, `
const handle = setTimeout(() => {}, 1);
if (typeof handle !== 'number' || handle < 1 || setTimeout(() => {}, 1) !== handle + 1) { throw new Error('timer handle is ' + typeof handle) }
for (const name of ['setTimeout', 'require', 'console']) {
  const value = globalThis[name];
  if (Object.getOwnPropertyNames(value).some((key) => !['length', 'name', 'log', 'info', 'warn', 'error', 'debug', 'clear'].includes(key))) {
    throw new Error(name + ' exposes ' + Object.getOwnPropertyNames(value).join())
  }
}`)
}

func TestManifestMatchesEmbeddedAssets(t *testing.T) {
	parsed := readManifest()
	if parsed.Reference != "postman-sandbox@6.7.4" || len(parsed.Files) == 0 {
		t.Fatalf("manifest = %+v", parsed)
	}
	listed := map[string]bool{}
	for _, entry := range parsed.Files {
		content, err := assetFiles.ReadFile("assets/" + entry.File)
		if err != nil {
			t.Fatalf("%s: %v", entry.File, err)
		}
		sum := sha256.Sum256(content)
		if hex.EncodeToString(sum[:]) != entry.Sha256 || entry.Version == "" || entry.Name == "" {
			t.Errorf("%s does not match its manifest entry %+v", entry.File, entry)
		}
		listed[entry.File] = true
	}
	files, err := fs.Glob(assetFiles, "assets/*.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !listed[strings.TrimPrefix(file, "assets/")] {
			t.Errorf("%s is not listed in MANIFEST.json", file)
		}
	}
}

func TestLibraryVersionsMatchPostmanSandbox(t *testing.T) {
	want := map[string]string{
		"ajv": "6.12.5", "chai": "4.4.1", "lodash": "4.18.1", "lodash3": "3.10.2", "moment": "2.30.1",
		"crypto-js": "3.3.0", "tv4": "1.3.0", "xml2js": "0.6.2", "cheerio": "0.22.0", "csv-parse/lib/sync": "1.2.4",
		"backbone": "1.6.0", "assert": "2.0.0", "buffer": "6.0.3", "uuid": "6.7.4",
	}
	for _, entry := range readManifest().Files {
		if version, found := want[entry.Module]; found && entry.Version != version {
			t.Errorf("%s version = %s, want %s", entry.Module, entry.Version, version)
		}
	}
}
