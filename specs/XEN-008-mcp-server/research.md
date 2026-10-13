# Research: XEN-008 MCP Server

Every decision below is final. Facts come from this repository (paths cited) or from the MCP Go SDK documentation
(Context7 `/modelcontextprotocol/go-sdk`).

## R1 — MCP library

- **Decision**: `github.com/modelcontextprotocol/go-sdk` **v1.8.0**, the latest stable release per
  `go list -m -versions` on 2026-10-13.
- **Rationale**: It is the official SDK and provides everything this feature needs:
  - `mcp.NewServer`, plus `mcp.AddTool` with typed input, which infers the schema and validates arguments.
  - `mcp.NewStreamableHTTPHandler`, which has DNS-rebinding/localhost protection on by default. Since v1.8.0
    `DisableLocalhostProtection` is opt-out, and cross-origin protection is added by wrapping the handler.
  - `mcp.StdioTransport` and `mcp.StreamableClientTransport` for the bridge.
  - `mcp.ToolAnnotations` with `ReadOnlyHint`, `DestructiveHint`, `IdempotentHint` and `OpenWorldHint`.
  - Protocol versions from 2024-11-05 to 2026-07-28 are negotiated automatically.
- **Alternatives**: `mark3labs/mcp-go` is community-run and lags on the newer protocol versions. A hand-rolled
  JSON-RPC server is too large and too easy to get wrong.
- **Import rule**: only `internal/infra/server/mcp`, `internal/infra/mcpbridge` and the BDD test client may import
  it. This mirrors the Wails rule in CLAUDE.md.

## R2 — Transport, port and lifecycle

- **Decision**:
  - The server is an `http.Server` on `127.0.0.1:<port>` serving `/mcp` through `StreamableHTTPHandler`, in stateful
    mode so the session carries the client name.
  - The default port is **47821**. It is a fixed, unassigned, high port, configurable within 1024–65535.
  - `properties.MCPPortOverride` comes from `XENON_MCP_PORT`, and BDD uses `0` for an ephemeral port. The actual bound
    address is reported by `Status()`.
  - The server starts from a Wails service's `ServiceStartup`, as `binding/request.go:80` already does for
    `ServiceShutdown`, when `settings.aiAccess.enabled`. It stops in `ServiceShutdown`, which also denies pending
    approvals.
  - Settings save calls `ServeAIAccess.Apply(settings)`. That stops the server, rebinds and starts again, and returns
    within 2 s (US1 sc. 6). Open sessions on the old port end.
- **Request size**: the SDK default `MaxRequestBodyBytes` of 4 MiB is kept. Larger messages get 413 (FR-002). Large
  payloads, such as imports, go by file path instead.
- **Port busy**: `net.Listen` fails. Status becomes `error` with reason `port <n> is in use` (code `MCP-01409`), and
  `AIAccessStatusChanged` is emitted. The app keeps running.
- **Alternatives**: a random port written to a discovery file was rejected, because clients need a stable URL in their
  config. A Unix socket or named pipe only was rejected, because Streamable HTTP clients can't use it and it isn't
  cross-platform in the same way.

## R3 — Access control (no token)

- **Decision**:
  - Bind to 127.0.0.1 only.
  - Keep the SDK's localhost protection on.
  - Add middleware that rejects with 403 when the `Host` is not `127.0.0.1:<port>` or `localhost:<port>`, and when an
    `Origin` header is present but is not `http://127.0.0.1[:port]` or `http://localhost[:port]`.
  - Requests with no Origin, which is how CLI and desktop clients send them, are allowed.
  - There is no token, per the product owner's decision.
- **Rationale**: This blocks remote hosts and browser pages (DNS rebinding and cross-site requests). The accepted
  exposure is other processes of the same user, which can read the vault key from the keychain anyway.

## R4 — The stdio bridge

- **Decision**:
  - `cmd/xenon/main.go` checks `os.Args[1] == "mcp"` **before** building the Wails app, so the existing
    `SingleInstance` (router.go:100) never activates the GUI.
  - `mcpbridge.Run()` reads `aiAccess` from `<DataDir>/app.json`, read-only, using the same properties for `DataDir`.
  - If AI access is disabled it prints `Xenon AI access is off — turn it on in Xenon → Settings → AI access (MCP).`
    and exits 1.
  - Otherwise it connects `mcp.StreamableClientTransport{Endpoint: http://127.0.0.1:<port>/mcp}`. If the connection is
    refused it prints `Xenon is not running — start Xenon with AI access on, then retry.` and exits 1.
  - Once connected, it pumps JSON-RPC messages both ways between `mcp.StdioTransport` and the client connection until
    stdin closes. It forwards `clientInfo.name` from the client's `initialize` unchanged.
  - The bridge never starts Xenon.
- **Config shown in Settings**: `{"command": "<os.Executable()>", "args": ["mcp"]}`, and for HTTP
  `{"type": "http", "url": "http://127.0.0.1:<port>/mcp"}`. Each is wrapped in the common `mcpServers.xenon` shape.

## R5 — Tool design and token efficiency

- **Decision**: 24 resource tools with `action`, split by effect so that each tool's hints are true for every action:
  - 12 read-only.
  - 10 that change data. 2 are W (`workspace_edit`, `settings_edit`). 8 are D because at least one action can remove
    data. That follows the contract's definition: an item, a variable or vault value, an overwritten file, a replaced
    list, or local changes. Changing a field in place is not removal. `send` is also open-world.
  - 2 destructive-only: `delete` and `git_push`.

  The full catalog is in [contracts/mcp-tools.md](contracts/mcp-tools.md).
- **Output encoding**: one compact JSON **text** content block per result. No `structuredContent` or output schema is
  emitted, because the SDK would serialize the payload twice and double the tokens.
  - Field names are short snake_case.
  - Empty, zero and default fields are omitted.
  - Lists default to 100 rows with a `next` cursor (an opaque offset).
- **Tool descriptions**: one sentence each. The input schema describes `action` as an enum, so the AI picks on the
  first try. SC-005 is asserted by a BDD step that counts the `tools/list` payload size: 6,000 tokens or fewer, at
  roughly 4 characters per token, so 24,000 characters or fewer.
- **Body cap**: text bodies inline up to 16,384 bytes of the masked text, with `truncated: true` and `size` (masked size; raw byte count for binary).
  `response {action: "body", offset, length≤65536}` reads more. Binary bodies, as classified by the existing response
  model, are never inlined; `export {action: "response_body", path}` saves them.

## R6 — Leak guard

- **Decision**: An application service `GuardAIOutput` holds a session-scoped `KnownSecrets` set.
  - Before each tool result is returned, it unions:
    - the active workspace's vault secrets and OAuth tokens (`collect_known_secrets` → the vault persistence
      `SecretsFor` and the OAuth token read);
    - every value already in the set (values seen this app session are never forgotten);
    - the forms registered by sends: the `send_request` service registers the `VariableContext.SecretValues()` of each
      send plus the derived `Authorization` basic value (base64 of `user:password`) when a secret took part.
  - Each value expands to: raw, `url.QueryEscape`, `url.PathEscape`, JSON-escaped (the inside of `json.Marshal`
    quotes), and standard base64. This reuses `SnippetSecretForms` (`entity/snippet.go:90`) and adds base64.
  - Masking is longest-first with `usecase.MaskSecrets` (`usecase/mask_secrets.go:11`) over the serialized JSON text.
    Values of 4 or more characters are replaced anywhere. Shorter values are replaced only when they make up a whole
    JSON string value.
  - The set lives only in memory and is never logged: its `LogValue` prints counts only.
- **Rationale**: Masking the final serialized text is the only placement that also covers fields nobody thought to
  mask, such as echoed bodies, test names, script errors and cookie values.
- **Mask before slicing**: body text is masked whole **before** the 16 KB cap or an offset/length slice is applied.
  Otherwise a secret straddling a cut point would come back as two unmasked halves. The final pass over the
  serialized text still runs.
- **Files written on the AI's behalf are AI output**: `export response_body` and `tree_edit create example
  from_last_response` write the **masked** text. A binary body that contains a known secret form is refused with
  `MCP-02403`. Without this, an echoed secret would land in a file the AI can read with its own tools, or in workspace
  YAML that git sync pushes. The same operations started by the human in the UI are unchanged (FR-021).
- **Accepted residual risks** (recorded in the spec Context, the Settings note and the PRD):
  1. A script that transforms a secret before it is output.
  2. An AI that sends `{{secret}}` to a destination it controls and reads it there. This is inherent to "AI can send
     requests that use secrets" (issue #3). Mitigation: "Confirm before send" environments and the AI client's own
     approval of `send`, which is hinted destructive and open-world.
- **Vault locked** (passphrase mode, not unlocked): no secret values are available to resolve or to leak, so sends
  resolve nothing (existing `SND-06409`). Values seen earlier in the session stay masked.

## R7 — Refusals (secret paths)

The tool layer refuses these before calling the controller, with error `MCP-02403` and the message
`"<what> is human-only — do it in Xenon."`:
- `snippet.generate` with `include_secrets: true`.
- `export.run` with `include_secrets: true`.
- `env_edit.set_vars` that turns a secret row plain. This is detected by comparing with the current table: any row
  where `secret` goes from true to false is refused, and `confirm_reveal` is never forwarded.
- Any vault or reveal operation. These have no tool, so the refusal is structural.

`secret_value` arguments are accepted only on `env_edit` (`add_var`, `set_vars` rows with `secret: true`). The
guard's arguments masking also hides them in History action rows.

## R8 — Approvals

- **Fact**: the confirm-before-send check already lives in the backend. `send_request.go:142-145` →
  `RejectUnconfirmedSend` returns 409 `SND-05409` with the environment's id, name and colour. It is bypassed by
  `confirmed_environment_id` (`dto/send_request.go:8,23`).
- **Decision**: the `send` tool calls the send controller without acknowledgement.
  - On `SND-05409`:
    - If `skipApprovals` is on, it re-calls at once with `confirmed_environment_id`.
    - Otherwise it calls `ApproveAISend.Request(...)`. The service:
      1. creates an `AIApproval`;
      2. calls the `FocusWindow` port (an infra adapter: Wails `window.Restore()` + `Focus()`);
      3. publishes `ai:approval-requested`;
      4. waits on a channel with `timeutils` `AfterFunc(120s)` (mockable through `mock/time.go`'s `AfterFunc`).
    - The UI answers through `AIAccess.AnswerApproval{id, approved}`.
    - On approve, the tool re-calls with `confirmed_environment_id`. Otherwise it returns `MCP-03403 denied by the user
      in Xenon` or `MCP-04408 approval timed out`.
  - Approvals are FIFO, and the UI shows one prompt at a time with a queue count.
  - `ServiceShutdown` resolves every pending approval as denied. When an MCP session closes, its pending approvals
    resolve as denied and their prompts are removed (FR-027).
- **Missing-secret and expired-token acknowledgements**: the `send` tool accepts `send_unresolved: true` →
  `acknowledge_missing_secrets`, and `send_expired_token: true` → the existing expired-token acknowledgement field.
  These are pass-throughs (FR-029).

## R9 — UI refresh after MCP writes

- **Fact**: Xenon records the hash of its own writes, and `handle_file_change.go:81` drops them. `TreeChanged` is
  published only by sync, import and the watcher. Environment, variable, vault and cookie services already publish
  their events.
- **Decision**: after each successful write or destructive tool call, the MCP pipeline calls
  `NotifyAIChange.Notify(localId, changes)`. `changes` is derived per tool and action, as listed in
  contracts/mcp-tools.md:
  - `tree` → `WorkspaceTreeChanged`.
  - A request save → also `RequestChangedOnDisk{requestId}`, which uses the UI's existing dirty-tab handling (US3
    sc. 2).
  - `env` → `VariablesChanged`, unless the service already emits it; duplicates are harmless.
  - `history` → `HistoryChanged`, a new event emitted for every recorded entry.
  - `cookies` → `CookiesChanged`.
  - `git` → `GitStatusChanged`.
  - `workspace` → `WorkspacesChanged`, a new event, whose payload carries the active local id so the window follows a
    switch.

  The frontend subscribes in `app/bootstrap.ts`.

## R10 — History: source, kind, per-source caps, filter

- **Facts**: history is `<DataDir>/workspaces/<localId>/history.jsonl`, appended per send and compacted at
  `entity.HistoryCap = 1000` (`entity/history.go:4,16-20`; `persistence/history.go`).
- **Decisions**:
  - Each JSON line gains `kind` (`send|action`), `source` (`user|ai`), `client`, and for action entries
    `activity{tool, action, target{id,name,path}, outcome, code, duration_ms, args}`. Missing fields read as
    `kind=send, source=user` (FR-032).
  - Compaction keeps the newest `historyUserCap` entries with source user and the newest `historyAiCap` entries with
    source ai (sends and actions together), merged by time.
  - Compaction runs on every append, as today. A read applies the caps in memory, so lowering a cap shows the effect
    at once and the file is trimmed on the next append (spec US6 sc. 8).
  - `List(source)` and `Clear(source)`: clearing a single source rewrites the file atomically with the other source's
    lines (`fileutils.WriteFileAtomic`); `all` truncates.
  - Every MCP call made while a workspace is active is appended after the call:
    - **Sends** are recorded by the send service with `SendOrigin{ai, client}`. They are **not** recorded again as
      action rows.
    - **Other tools** become action rows, recorded by `RecordAIActivity`, which never returns an error to the caller.
  - Calls with no active workspace are logged only (US6 sc. 6).
  - `workspace_edit.switch` is recorded in the workspace that is active once the call finishes.
  - Action row `args` hold the call's arguments, JSON-serialized and capped at 2,048 bytes with `…` appended.
    Secret-carrying fields are first redacted **by name**, whatever the call's outcome:
    - `secret_value` anywhere, including `vars[].secret_value` and the `variables` of `tree_edit save_settings`;
    - `value` in `env_edit` `add_var` and `set_current`, always, whether or not the row is secret;
    - `value` of any object that also has `secret: true`;
    - `passphrase` anywhere.

    Redacting by tool and field name (instead of by looking the row up) keeps the redactor a pure use case and is
    safe when the lookup itself would fail.

    Only then does the leak guard run. This ordering is needed because a failed call (for example `VLT-04409`, vault
    locked) never puts the value in the vault, so the guard alone would not know it and would write plaintext to
    `history.jsonl`.
  - The filter value is per-workspace UI state, added to the existing `ui_state` (`entity/ui_state.go`) as
    `historySource`. The default is `all`.

## R11 — Settings as a tab

- **Facts**:
  - Today Settings is `features/settings/components/SettingsDialog.tsx`, a Dialog with a Save form for the script
    timeout, plus `GitAccounts` and `ShortcutsList`.
  - It is opened by `SettingsButton` and the `settings` hotkey in `app/hotkeys.ts`, including when no workspace is
    open (`NO_WORKSPACE_HOTKEYS`, `App.tsx:119`).
  - Environment tabs (`features/tabs`, `EnvironmentTabTarget`) already implement a dirty dot, Save / Don't Save /
    Cancel, and restore.
- **Decision**:
  - Add tab kind `settings`: a singleton, always pinned, reusing the environment-tab dirty and close-prompt machinery.
    The tab target is persisted, but its draft is not.
  - The store keeps one draft of all form fields, and Save calls `SettingsService.Save` once. Git account actions keep
    calling their own bindings immediately.
  - With no workspace open, `App.tsx` renders `SettingsTab` as a full main-area page with a close button.
  - `SettingsDialog.tsx` and its test are deleted.

## R12 — Activity in History vs. the app log

Every MCP call is also logged at Info level with the client, tool, action, outcome and duration. Arguments are omitted
from the log, because `logger.WithRedactions` covers only known secrets. The log line is the fallback when no
workspace is active.
