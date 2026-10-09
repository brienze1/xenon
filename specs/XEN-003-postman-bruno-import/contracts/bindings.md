# Contract: Go ↔ UI bindings — stage 3 additions

Conventions as in XEN-001/XEN-002. New prefixes: `IMP` import · `EXM` examples.

## 1. `ImportService`
| Method | Input | Output | Notes / errors |
|---|---|---|---|
| `Detect(req)` | `{ paths?: string[], text?: string }` | `ImportPreview { source_id, kind, version, items: [{ key, kind, name, counts, warnings_count }], warnings[] }` | unsupported format `IMP-00415` (message names the format, e.g. "Postman v3 YAML is not supported yet"); malformed `IMP-01400`; zip too large `IMP-02413`; zip-slip `IMP-03400` |
| `PickSources()` | — | `{ paths[] }` | native file/folder picker (multi-select, zip/json/folders) |
| `Clashes(req)` | `ImportPlan` | `{ clashes: Clash[] }` | lets the UI ask before importing |
| `Import(req)` | `ImportPlan` | `ImportResult` | unresolved clashes `IMP-04409` (details list them); unknown source `IMP-05404`; cancellable via ctx until writing starts |
| `Discard(req)` | `{ source_id }` | — | frees temp files |
| `ParseCurl(req)` | `{ text }` | `{ request: Request, notices: string[] }` | used by the URL bar paste; `IMP-06400` not a cURL command; `IMP-07400` non-HTTP URL |

Events: `import:progress { source_id, done, total, current_item }` (every item, and at least every
250 ms during long items).

## 2. `ExampleService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Get(req)` | `{ local_id, request_id, example_id }` | `Example` (request + response DTO; body inline ≤ 1 MB, else `body_url` via the XEN-001 response route keyed `example:<id>`) | `EXM-01404` |
| `CreateFromResponse(req)` | `{ local_id, request_id, tab_id, name }` | `Example` | uses `RequestService.LastResult` + the tab's sent request (unresolved); `EXM-02404` no response |
| `Create(req)` | `{ local_id, request_id, name }` | `Example` | empty example (copy of the request, 200 OK, empty body) |
| `Save(req)` | `{ local_id, request_id, example, base_revision, force }` | `Example` | revision conflict `REQ-02409` (same rule as requests) |
| `Rename(req)` / `Duplicate(req)` / `Delete(req)` | `{ local_id, request_id, example_id, name? }` | `Example` / `Example` / — | duplicate → "<name> Copy" after original |

## 3. Changes to earlier services
- `TreeService.GetTree` nodes: requests carry `examples: [{ id, name, code }]` (kind `example`).
- Tabs (UI state) gain `kind: request | example` and `example_id`.
- `RequestService` responses: `Request.auth.type` may be `unsupported` (read-only data shown).
