# Research: Scripts, Tests and Console (Stage 4)

Verified 2026-10-09 by reading goja, goja_nodejs, grafana/sobek, postman-sandbox 6.7.4,
postman-collection 5.3.1 and chai-postman 2.0.1 sources, and by a Go harness that ran chai,
lodash, moment, crypto-js, ajv and tv4 inside goja.

## R1. JavaScript engine
- **Decision**: `github.com/dop251/goja` pinned by pseudo-version
  (`v0.0.0-20261009133625-a4bedf5f6000`) + `github.com/dop251/goja_nodejs`
  (`v0.0.0-20260918173711-b481721df8a2`) for `require`, `console` base and the event loop.
  Upgrades are deliberate (bump + parity suite).
- **Why not grafana/sobek**: its addition is experimental ESM, which Postman scripts don't use;
  upstream goja gets fixes first.
- **Language support**: ES5.1 + nearly all ES2015–ES2022 (classes incl. private fields,
  async/await, Promise, generators, `?.`, `??`, BigInt, lookbehind, named groups). Missing:
  async generators, ES modules, WeakRef/FinalizationRegistry, SharedArrayBuffer/Atomics, some
  RegExp flags (`v`, `d`, `\p{}` unreliable), Temporal, `Intl`. These produce a clear
  "not supported in Xenon" error (spec Edge Cases) except `Intl` (R6).

## R2. Runtime model
- **Decision**: **fresh `goja.Runtime` per script** (Postman resets scope per script; reusing a
  runtime leaks declarations — verified). Library sources are compiled once into
  `*goja.Program` and cached process-wide; loading chai+lodash+moment costs ~25 ms fresh, so
  libraries are loaded **lazily** on first `require`/global access (globals defined as lazy
  getters).
- Each script is wrapped like postman-sandbox: `;(async()=>{ <code> })()` to allow top-level
  `await`; completion = wrapper settled **and** event loop has no pending timers /
  sendRequest / cookie-jar calls (goja_nodejs `eventloop.Run`).
- Unhandled rejections via `SetPromiseRejectionTracker` → script error.

## R3. Limits
- **Timeout**: `time.AfterFunc(limit)` → `vm.Interrupt` + `loop.Terminate()`; error
  `script_timeout`.
- **Stack**: `SetMaxCallStackSize(10_000)` → `script_stack_overflow`.
- **Memory**: goja has no limit. A guard polls `runtime/metrics` (`/memory/classes/heap/objects:bytes`)
  every 50 ms while a script runs; growth > 512 MB since script start → interrupt with
  `script_memory`. Process-wide measure, acceptable because scripts of one send run sequentially
  and the guard only aims to keep the app alive.

## R4. pm API implementation split
- **Go-implemented** (behaviour and data owned by Go per constitution III): variable scopes
  (call application use cases from XEN-002), `pm.sendRequest` / `pm.execution.runRequest`
  (call the send service), `pm.cookies` (response cookies; jar → stage 5), `pm.vault`, `pm.info`,
  `pm.execution.skipRequest/location`, console sink, test recorder, `crypto` (Go crypto),
  `TextEncoder/Decoder`, `atob/btoa`, `structuredClone` (via JSON + Map/Set/Date handling),
  `URL`/`URLSearchParams` (Go `net/url` semantics aligned to WHATWG for common cases).
- **JS shims** (embedded, small, readable): `pm.request`/`pm.response` object model mirroring
  postman-collection (`Url`, `PropertyList`, `HeaderList`, `RequestBody`, `Response` methods)
  backed by plain data from Go; chai-postman assertions ported to the shim's objects (~15
  helpers, messages copied); legacy interface (`postman.*`, `tests`, `responseBody`, …) mapped to
  pm; `pm.test` runner semantics from `pmapi-setup-runner.js`.

## R5. Libraries (versions = postman-sandbox 6.7.4)
| Library | Version | Delivery |
|---|---|---|
| chai | 4.4.1 | vendored UMD (`global` shim) |
| lodash / lodash3 (legacy `_`) | 4.18.1 / 3.10.2 | vendored |
| moment | 2.30.1 | vendored min build |
| crypto-js | 3.3.0 | vendored |
| ajv | 6.12.5 | vendored `dist/ajv.bundle.js` |
| tv4 | 1.3.0 | vendored |
| uuid | vendored (Postman's) | vendored, random source from Go `crypto/rand` |
| xml2js | 0.6.2 | esbuild bundle with Node shims |
| cheerio | 0.22.0 | esbuild bundle with Node shims |
| csv-parse (sync) | 1.2.4 | esbuild bundle |
| backbone | 1.6.0 | vendored |
| postman-collection | 5.3.1 | **classes re-exported from the JS shims** (Url, Header, PropertyList, Request, Response, Variable, VariableScope) rather than bundling the full SDK |
| Node modules | assert 2.0.0, buffer 6.0.3, events, path, punycode, querystring, string_decoder, stream, timers, url, util | esbuild-bundled browser builds (`assert`, `buffer`, `events`, `path-browserify`, `punycode`, `querystring-es3`, `string_decoder`, `stream-browserify`, `timers-browserify`, `url`, `util`) |
- All JS assets live in `internal/integration/webservice/sandbox/assets/`, built by
  `tools/sandbox-assets/` (package.json with exact versions + esbuild script), committed with a
  `MANIFEST.json` (name, version, sha256), embedded via `go:embed`. CI rebuilds and fails on
  drift.

## R6. Intl
- **Decision**: FormatJS polyfills (`@formatjs/intl-getcanonicallocales`, `intl-locale`,
  `intl-pluralrules`, `intl-numberformat`, `intl-datetimeformat` with `en-US` locale data and UTC +
  local time zone data) bundled into the assets and installed lazily on first `Intl` access;
  `toLocaleString`/`toLocaleDateString` patched to use them. Other locales → en-US + one console
  warning per script.

## R7. Parity oracle
- **Decision**: `tools/parity/` — a Node script that runs each corpus case through
  **postman-sandbox 6.7.4** with the case's request, response and variable fixtures and writes the
  observed outcome (variable mutations, sent request after pre-request mutations, test results,
  console calls, errors). A Go test runs the same case through Xenon's sandbox and diffs. Corpus:
  `test/integration/resources/scripts/corpus/*.json` (≥ 200 cases: hand-written per API in FR-008/9/13/14
  plus real-world scripts from public Postman collections with their licences noted). Runs in CI
  (Node is already there for the frontend). SC-001/SC-005.

## R8. Console transport
- **Decision**: Go `ConsoleService` holds a per-session ring buffer (5,000 entries) and emits
  `console:entries` events in batches every 50 ms (or 100 entries). Masking is applied when an
  entry is created, against the active workspace's secret values (XEN-002 vault, unlocked).
  Bodies > 1 MB are referenced (`body_url` via the XEN-001 response route, key `console:<id>`).

## R9. Editor support
- **Decision**: the pm API reference (members, signatures, descriptions) is a JSON document
  generated from the Go/JS shim definitions at build time and served by
  `ScriptService.Reference()`; snippets list from Go too. Syntax checking via
  `ScriptService.Check(text)` (goja `Compile`), debounced 300 ms. CodeMirror `@codemirror/lang-javascript`
  + custom completion source.
