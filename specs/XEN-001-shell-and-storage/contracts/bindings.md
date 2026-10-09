# Contract: Go ↔ UI bindings (stage 1)

The UI talks to Go only through these Wails-bound services (wrapped by
`frontend/src/platform/services/*.ts`) and the events in §8. Every method's first Go parameter is
`ctx context.Context` (omitted in TS; `promise.cancel()` cancels it). DTO JSON fields are
snake_case (dto skill). IDs are ULID strings.

**Errors**: a failed call rejects with `RuntimeError` whose `cause` is
`{ "error": { "id", "code", "description", "error_details": [{ "attribute", "messages" }] } }`.
`code` follows `{PREFIX}-{NN}{STATUS}`; the UI branches on `code`, never on text. Prefixes:
`WKS` workspaces · `TRE` tree items · `REQ` requests · `SND` sending · `HIS` history ·
`UIS` UI state · `DLG` dialogs · `UNK` unknown.

## 1. `WorkspaceService`
| Method | Input | Output | Notes / errors |
|---|---|---|---|
| `List()` | — | `WorkspaceList { workspaces: [{ local_id, name, path, missing }], active_local_id }` | |
| `Create(req)` | `{ name, path }` | `WorkspaceSummary` | path must be empty or not exist (`WKS-01409` not empty), not overlap a registered one (`WKS-02409`); name 1–200 (`WKS-00400`) |
| `Open(req)` | `{ path }` | `WorkspaceSummary` | needs `workspace.yaml` (`WKS-03404`), overlap `WKS-02409`, unsupported version `WKS-04422` |
| `Rename(req)` | `{ local_id, name }` | `WorkspaceSummary` | writes `workspace.yaml` name |
| `Remove(req)` | `{ local_id, move_to_trash }` | `WorkspaceList` | trash failure `WKS-05500`; picks next active |
| `Activate(req)` | `{ local_id }` | `WorkspaceSnapshot { summary, tree, ui_state }` | folder missing `WKS-06404`; starts watcher on it, stops previous |
| `Locate(req)` | `{ local_id, path }` | `WorkspaceSummary` | same checks as Open |

## 2. `TreeService`
`TreeNode { id, kind (collection|folder|request|error), name, method?, url?, children?, error? }`
| Method | Input | Output | Notes |
|---|---|---|---|
| `GetTree(req)` | `{ local_id }` | `TreeNode[]` | collections in order; error nodes carry `error` text (FR-016) |
| `Search(req)` | `{ local_id, query }` | `{ match_ids: string[], ancestor_ids: string[] }` | case-insensitive name/URL substring, requests only |
| `CreateCollection(req)` | `{ local_id, name }` | `TreeNode` | appended to workspace order |
| `CreateFolder(req)` | `{ local_id, parent_id, name }` | `TreeNode` | parent = collection/folder (`TRE-01404`) |
| `CreateRequest(req)` | `{ local_id, parent_id, name, request? }` | `TreeNode` | `request` = draft from an unsaved tab (Save) |
| `Rename(req)` | `{ local_id, id, name }` | `TreeNode` | renames file/dir, updates parent order |
| `Duplicate(req)` | `{ local_id, id }` | `TreeNode` | new ids for the copy and all descendants, "<name> Copy", placed after original |
| `Delete(req)` | `{ local_id, id }` | `{ deleted_request_ids: string[] }` | recursive; UI closes those tabs (FR-022) |
| `GetSettings(req)` | `{ local_id, id }` | `ItemSettings { id, kind, name, description, auth, headers }` | collection/folder only |
| `SaveSettings(req)` | `ItemSettings` | `ItemSettings` | |

## 3. `RequestService`
`Request` DTO mirrors contracts/workspace-format.md §5 (snake_case: `path_variables`,
`auto_headers.disabled`, `settings.timeout_ms`, …) plus read-only `revision` (content hash).
| Method | Input | Output | Notes |
|---|---|---|---|
| `Get(req)` | `{ local_id, id }` | `Request` | `REQ-01404` not found |
| `Save(req)` | `{ local_id, request, base_revision, force }` | `Request` | `REQ-02409` when the file changed since `base_revision` and `force` is false (FR-017 "Keep mine" = retry with `force: true`) |
| `SaveAs(req)` | `{ local_id, request, parent_id, name }` | `Request` | new id |
| `ApplyUrl(req)` | `{ request }` | `Request` | re-derives `params` (keeping disabled rows/descriptions) and `path_variables` from `url` |
| `ApplyParams(req)` | `{ request }` | `Request` | rebuilds `url` query from enabled `params` and path from `path_variables` keys |
| `ParseBulk(req)` / `FormatBulk(req)` | `{ text }` / `{ rows }` | `{ rows }` / `{ text }` | bulk edit (`key:value`, `//` = disabled) |
| `AutoHeaders(req)` | `{ local_id, request }` | `[{ key, value, disabled }]` | effective auto headers incl. inherited auth/headers preview |
| `Send(req)` | `{ local_id, tab_id, request }` | `SendResult` | cancellable via ctx; always appends history; see below |
| `ReleaseResponse(req)` | `{ tab_id }` | — | frees memory/temp file on tab close |
| `SaveResponseBody(req)` | `{ tab_id, path }` | — | exact bytes (`SND-03404` no response) |

`SendResult { tab_id, ok, status_code, status_text, proto, headers: [{key,value}], size_bytes,
content_type, is_binary, body_url, inline_body?, timings: { dns_ms, connect_ms, tls_ms,
first_byte_ms, download_ms, total_ms }, error?: { kind, message, hint } }`.
`inline_body` is set when the body ≤ 1 MB and is text; otherwise the UI fetches `body_url`
(`/xenon/response/{tab_id}`, §6). `error.kind` ∈ `dns | connection_refused | timeout | tls |
cancelled | missing_file | invalid_url | other`; `hint` names the setting that changes it.

## 4. `HistoryService`
| Method | Input | Output |
|---|---|---|
| `List(req)` | `{ local_id, query, limit, before_id? }` | `[{ id, sent_at, method, url, status_code?, error_kind?, total_ms, size_bytes }]` newest first |
| `Get(req)` | `{ local_id, id }` | `{ entry, request }` (request opens in a new unsaved tab) |
| `Clear(req)` | `{ local_id }` | — |

## 5. `UIStateService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Get(req)` | `{ local_id }` | `UIState` (data-model.md WorkspaceUIState) | defaults when absent |
| `Save(req)` | `UIState` | — | UI debounces (300 ms); written atomically |

## 6. `ResponseBodyRoute` (HTTP handler service, `Route: "/xenon/response"`)
`GET /xenon/response/{tab_id}` → `200` with the exact body bytes and the response's
`Content-Type`; `404` when there is no body for that tab. Same-origin, served by the Wails asset
server; not reachable from outside the app.

## 7. `DialogService` (infra-only helper, native dialogs)
| Method | Output |
|---|---|
| `PickFolder({ title })` | `{ path }` (empty = cancelled) |
| `PickFile({ title })` | `{ path }` |
| `PickSaveFile({ title, suggested_name })` | `{ path }` |

## 8. Events (Go → UI, typed via `application.RegisterEvent`)
| Name | Payload | When |
|---|---|---|
| `workspace:tree-changed` | `{ local_id }` | files added/removed/renamed outside Xenon → UI refetches tree |
| `request:changed-on-disk` | `{ local_id, request_id, deleted, revision }` | a request file changed/deleted outside Xenon (FR-017) |
| `workspace:folder-missing` | `{ local_id }` | active workspace folder disappeared |
