# Data Model: Shell and Storage (Stage 1)

Two kinds of data:

1. **Workspace data**: YAML in the workspace folder, shareable via git. The full schema is in
   [contracts/workspace-format.md](contracts/workspace-format.md).
2. **Local data**: per-machine JSON in Xenon's app-data folder, never in a workspace folder,
   never synced (FR-005, FR-038c).

Domain entities live in `internal/domain/entity/` and enums in `internal/domain/enums/` (see
plan.md for the package map). Field names below are domain names; the YAML/JSON names are in
the contracts.

## Workspace data entities

### Workspace
| Field | Type | Rules |
|---|---|---|
| ID | ULID | generated on create; immutable |
| Name | string | 1–200 chars, trimmed, any Unicode |
| Description | markdown string | optional |
| Order | []string | collection directory names |
| Variables | []KeyValue | reserved (stage 2 globals) |
| Sync | map | reserved (stage 7) |
| Unknown | YAML node | preserved verbatim |

Relationships: has many **Collection**s (via `collections/` directories).

### Collection / Folder (same shape, `Kind` distinguishes)
| Field | Type | Rules |
|---|---|---|
| ID | ULID | immutable |
| Kind | enum `collection` \| `folder` | |
| Name | string | 1–200 chars |
| Description | markdown | |
| Order | []string | child dir/file names |
| Auth | Auth | `inherit` on a collection resolves to no auth |
| Headers | []KeyValue | applied to all descendants unless overridden |
| Variables | []KeyValue | reserved (stage 2); collection only — folders have no variables |
| Scripts | Scripts | reserved (stage 4) |
| InheritScripts | bool | folder only, default true (stage 4) |
| Postman | YAML node | reserved passthrough (stage 3) |

Relationships: Collection has many Folders and Requests; Folder has many Folders and Requests;
nesting unlimited.

### Request
| Field | Type | Rules |
|---|---|---|
| ID | ULID | immutable |
| Name | string | 1–200 chars; duplicates among siblings allowed |
| Description | markdown | |
| Method | string | one of the 7 standard methods or a custom token matching `^[A-Z][A-Z0-9_-]{0,31}$` (stored upper-case) |
| URL | string | free text; may contain `{{var}}` and `:pathVar` |
| Params | []KeyValue | invariant: enabled rows, in order, equal the URL's query |
| PathVariables | []PathVariable | invariant: one per `:name` segment of the URL path, in order |
| Headers | []KeyValue | |
| DisabledAutoHeaders | []string | names from the auto-header set (case-insensitive) |
| Auth | Auth | |
| Body | Body | |
| Settings | RequestSettings | |
| Scripts / InheritScripts / Postman | reserved | stages 3–4 |

### Value objects
- **KeyValue**: Key, Value, Enabled (default true), Description. `Secret` reserved (stage 2).
- **PathVariable**: Key, Value, Description.
- **Auth**: Type (enum `inherit|none|bearer|basic|apikey|oauth2|awsv4|digest`) plus per-type
  sub-structs (all preserved). Stage 1 sends only bearer/basic.
- **Body**: Mode (enum `none|formdata|urlencoded|raw|binary|graphql`) plus all per-mode sub-structs
  (preserved across mode switches). `RawLanguage` enum `json|xml|text|html|javascript`.
  `FormDataRow` = KeyValue + Type (`text|file`) + Src.
- **FileRef** (`Src`): stored path; resolver turns it into an absolute OS path: relative → joined
  with the workspace folder (`/` → OS separator); absolute → as-is. State `missing` when it does
  not exist at send time.
- **RequestSettings**: TimeoutMs (≥ 0, default 0 = none), FollowRedirects (default true),
  VerifySsl (default true), EncodeUrl (default true).

### Effective request (derived, not stored)
Built at send time from the request and its ancestors (FR-014):
- **Auth**: walk Request → Folder… → Collection; first non-`inherit` wins; collection `inherit` = none.
- **Headers**: start from the collection's enabled headers, override by name (case-insensitive)
  with each folder's going down, then the request's; request wins.
- **Auto headers**: computed from body/auth/URL (Content-Type by body mode/raw language,
  Content-Length, Host, `User-Agent: Xenon/<version>`, `Accept: */*`,
  `Accept-Encoding: gzip, deflate, br`, `Connection: keep-alive`, Authorization from auth) minus
  `DisabledAutoHeaders`, and minus any name the user set explicitly (user header wins).
- **URL**: path variables substituted; when EncodeUrl is on, path and query are percent-encoded
  where not already encoded.

## Local data (app data)

Root: `os.UserConfigDir()/Xenon/` →
macOS `~/Library/Application Support/Xenon/`, Windows `%AppData%\Xenon\`,
Linux `$XDG_CONFIG_HOME/xenon/` (default `~/.config/xenon/`).

```
Xenon/
├── app.json                          ← registry + global prefs
└── workspaces/
    └── <localId>/                    ← one per registered workspace
        ├── state.json                ← UI state (tabs, drafts, sidebar, layout)
        └── history.jsonl             ← send history (newest appended)
        # stage 2: vault.age · stage 5: cookies.json
```

All files are written atomically. Every file carries `"version": 1`. Because drafts and
history can contain values the user typed (e.g., a Bearer token), the app-data directories are
created owner-only (0700) and files 0600 on macOS/Linux; on Windows they inherit the per-user
`%AppData%` ACL.

### AppRegistry (`app.json`)
| Field | Type | Rules |
|---|---|---|
| Version | int | 1 |
| Workspaces | []RegisteredWorkspace | order = switcher order |
| ActiveLocalID | string | last active; empty → welcome screen |
| Window | WindowState | size/position/maximized |

**RegisteredWorkspace**: LocalID (ULID generated at registration — *not* the workspace's own
`id`, because two clones of the same repo may be registered from different folders), Path
(absolute, cleaned), DisplayNameCache (last-known name for the switcher when the folder is
missing), AddedAt.

Validation: Path unique; no registered path may be inside or contain another registered path
(edge case "two workspaces, same folder").

### WorkspaceUIState (`state.json`)
| Field | Type |
|---|---|
| Version | int |
| Tabs | []Tab (in strip order) |
| ActiveTabID | string |
| PreviewTabID | string (at most one preview tab) |
| Sidebar | Expanded []ItemID, SelectedID, View (`collections\|history`), Search string |
| Layout | SidebarWidth, SidebarVisible, RightPanelWidth, RightPanelVisible, RightPanelView (`info`), ConsoleHeight, ConsoleVisible, Split (`stacked\|side-by-side`), SplitRatio (0.2–0.8) |

**Tab**: TabID (ULID), RequestID (ULID or empty for new tabs), Draft (full Request snapshot when
there are unsaved edits or for unsaved new tabs; absent when clean), BaseRevision (hash of the
file content the draft started from — used for FR-017 conflict detection), Pinned bool.

Tab title is derived (request name or "Untitled Request"), never stored.

### HistoryEntry (`history.jsonl`, one JSON object per line)
| Field | Type |
|---|---|
| ID | ULID (time-sortable) |
| SentAt | RFC 3339 timestamp |
| Request | full Request snapshot as written by the user (no variable resolution) |
| SourceRequestID | ULID or empty |
| Result | Status code + text, or ErrorKind + message; TotalMs; SizeBytes |

Cap: 1,000 newest. Appends are O(1); when the file exceeds 1,200 lines it is compacted to the
newest 1,000 (atomic rewrite). "Clear history" truncates the file.

## In-memory only

### Response
StatusCode, StatusText, Proto, Headers (ordered list, repeats kept), Body (bytes; streamed to a
temp file above 5 MB and referenced by path), SizeBytes (body, as received), ContentType,
Timings {DNS, Connect, TLS, FirstByte, Download, Total} in ms, or an **Error** with Kind enum:
`dns | connection_refused | timeout | tls | cancelled | missing_file | invalid_url | other`.

Kept per tab in the Go request service's memory (FR-038), dropped on tab close or app exit.
Temp files live under `os.TempDir()/xenon-<pid>/` (owner-only permissions) and are removed on
tab close and exit; at startup, stale `xenon-<pid>` directories of processes that are no longer
running are deleted.

## State transitions

### Tab
```
(none) --single click--> Preview --edit / double-click--> Pinned(clean)
Preview --single click other request--> Preview(other)            [same tab reused]
Pinned(clean) --edit--> Pinned(dirty) --save--> Pinned(clean)
Pinned(dirty) --close--> prompt: Save → saved+closed | Don't save → closed | Cancel → unchanged
NewTab(dirty, no RequestID) --save (pick location)--> Pinned(clean, RequestID set)
Any --request deleted--> closed (dirty → offer "keep as new tab" first)
Pinned(dirty) --file changed on disk--> Conflict banner: Keep mine | Load from disk
```

### Send
```
Idle --Send--> InFlight --response--> Done(response)
InFlight --Cancel--> Done(error: cancelled)
InFlight --failure--> Done(error: kind)
(Every transition to Done appends a HistoryEntry.)
```
