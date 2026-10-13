# Contract: MCP Tools (24)

Server: `http://127.0.0.1:<port>/mcp` (Streamable HTTP). Implementation name `xenon`, version = app version.

## Conventions (all tools)

- **Limits**: a request body over 4 MiB is rejected with HTTP 413. A rejected Host or Origin gets HTTP 403 and a
  warning log line, without the body.
- **Arguments**: `action` (enum, required where the tool has more than one action) plus action-specific fields. Ids
  are Xenon ids (ULIDs) as returned by list/get. Unknown actions or invalid arguments return `MCP-07400`.
- **Workspace**: every tool except `workspace` / `workspace_edit` acts on the active workspace. When none is active it
  returns `MCP-06404 No active workspace — create or open one`.
- **Result**: one text content block holding compact JSON. Field names are snake_case, and empty, zero and default
  fields are omitted. Lists return `{items:[…], next?:"<cursor>"}`, 100 per page. Pass `cursor` to continue.
- **Errors**: `isError: true` with text `{"code":"XXX-NNNNN","message":"…"}`. This is the same code and message as the
  Wails binding, masked (FR-025).
- **Pipeline per call** (`infra/server/mcp/call.go`):
  1. Resolve the client name and active workspace.
  2. Adapt the arguments to the controller DTO(s).
  3. Call the controller(s).
  4. Run the leak guard on every **whole** text value that the next step could cut, such as response bodies and
     long strings. A secret that straddles a cut point can then never come back as two unmasked halves.
  5. Compact the result: cap and slice, page, omit empty fields.
  6. Serialize it.
  7. Run `AIAccess.Guard` (the leak guard) again over the serialized text.
  8. Record the call (`AIAccess.Record`). A send that reached the send service is recorded there as a send row. Every
     other call is recorded here as an action row, including a `send` that was refused, denied, timed out or invalid
     before it went out.
  9. On a successful write, call `AIAccess.Notify`.
  10. Return.
- **Idempotent**: `idempotentHint:true` on every R tool, and on `settings_edit`, `cookies_edit` and `delete` (repeating them has no further effect). `false` on every other tool.
- **Hints**: R = `readOnlyHint:true`. W = `readOnlyHint:false, destructiveHint:false`. D = `readOnlyHint:false,
  destructiveHint:true`. A tool is D when **any** of its actions can remove data: an item, a variable or its vault
  value, a file it overwrites, a list it replaces, or local changes. Changing a field's value in place is not
  removal. `openWorldHint` is false except for `send`, `git_edit`, `git_push` and `tree_edit` (oauth
  actions).
- **Descriptions**: the exact one-sentence description is given per tool below.

Item kinds in `tree`: `collection`, `folder`, `request`, `example`.

---

## Read-only tools (12)

Hint summary: R = `workspace`, `tree`, `env`, `history`, `docs`, `response`, `snippet`, `cookies`, `console`, `git`, `scripts`, `settings`. W = `workspace_edit`, `settings_edit`. D = `tree_edit`, `env_edit`, `docs_edit`, `cookies_edit`, `git_edit`, `import`, `export`, `send`, `delete`, `git_push`.

### 1. `workspace` — R — "List workspaces or show the active one."
| action | args | returns |
| --- | --- | --- |
| `list` | — | `items:[{id,name,active?:true}]` |
| `active` | — | `{id,name,path,environment_id?}` or `MCP-06404` |
Controllers: `WorkspaceList.Handle` (its result marks the active workspace).

### 2. `tree` — R — "Browse, read and search collections, folders, requests and examples."
| action | args | returns |
| --- | --- | --- |
| `list` | `parent_id?` (omit → collections), `depth?` (1 default; `"all"` = full tree) | `items:[{id,kind,name,method?,children?}]` (children = count; nested `items` only when depth > 1) |
| `get` | `id` | the item: request → `{id,name,method,url,params,headers,body,auth,scripts?,description?,path}`; folder/collection → `{id,kind,name,description?,path,children}`; example → `{id,name,request,response}` |
| `settings` | `id` (collection/folder) | `{auth,headers,scripts,variables,description}` (secret variable values `•••••`) |
| `search` | `query`, `limit?` | `items:[{id,kind,name,method?,path}]` |
| `token` | `owner_id` | OAuth 2.0 token metadata `{type,expires_at,scopes,state}` (never the token) |
Controllers: `TreeSearch.GetTree`, `RequestGet`, `ExampleManage.Get`, `TreeSettings.GetSettings`, `TreeSearch`,
`OAuth.Get`.

### 3. `env` — R — "List environments and read or inspect variables."
| action | args | returns |
| --- | --- | --- |
| `list` | — | `items:[{id,name,color,active?:true,confirm_before_send?:true}]` |
| `get` | `scope`: `globals` \| `env:<id>` \| `collection:<id>` | `{name, vars:[{key,initial?,current?,secret?:true,state?,enabled?:false,description?}]}`: secret → `initial`/`current` omitted, `state` (`set`\|`missing`\|`locked`) |
| `inspect` | `text` (for example `{{baseUrl}}/users`) or `request_id` | `refs:[{name,scope,value}]` (secret → `•••••`) |
Controllers: `EnvironmentList.Handle`, `VariableTable.GetTable`, `VariableInspect.Inspect`.

### 4. `history` — R — "List and read History (sends and AI activity)."
| action | args | returns |
| --- | --- | --- |
| `list` | `source?` (`all` default \| `user` \| `ai`), `query?`, `cursor?` | `items:[{id,at,kind,source,client?, method?,url?,status?,ms?, tool?,action?,target?,outcome?}]` |
| `get` | `id` | a send → `{request (as written), status, ms, size, error?}`; an action → `{tool,action,target,outcome,code?,ms,args}` |
Controller: `History.List/Get` (source filter).

### 5. `docs` — R — "Read a collection's OpenAPI document or preview the plan for an edited one."
| action | args | returns |
| --- | --- | --- |
| `get` | `collection_id` | `{yaml}` |
| `plan` | `collection_id`, `yaml` | `{updated:[…],created:[…],deleted:[…]}` or a parse error with line/column |
Controllers: `TreeDocumentation.GetDocumentation/PlanDocumentation`.

### 6. `response` — R — "Read the last response of an AI send, including more of a truncated body."
| action | args | returns |
| --- | --- | --- |
| `last` | — | same shape as `send`'s result for this session's last send |
| `body` | `offset` (default 0), `length` (default 16384, max 65536) | `{text, offset, size, truncated}`; the whole body is masked first, then sliced, and `offset` and `size` refer to the masked text; binary → `{binary:true, content_type, size}` |
Each MCP session uses the tab id `mcp:<session-id>`. Controllers: `RequestResponse.LastResult/OpenBody`.

### 7. `snippet` — R — "Generate a code snippet for a request (secrets masked)."
| action | args | returns |
| --- | --- | --- |
| `languages` | — | `items:[{id,name}]` |
| `generate` | `request_id` or `request`, `language` | `{code}` — `include_secrets:true` → `MCP-02403` |
Controller: `Snippet.Languages/Generate` (always with `IncludeSecrets=false`).

### 8. `cookies` — R — "List cookies and the cookie allowlist."
| action | args | returns |
| --- | --- | --- |
| `list` | `domain?` | `items:[{domain,name,value,path,expires?,secure?,http_only?}]` |
| `allowlist` | — | `{domains:[…]}` |
Controller: `Cookie.List/GetAllowlist`.

### 9. `console` — R — "Read console entries from sends and scripts."
| action | args | returns |
| --- | --- | --- |
| `list` | `type?` (`request`\|`log`\|`error`), `query?`, `cursor?` | `items:[{id,at,type,summary}]` |
| `get` | `id` | the entry (already masked by the console, then by the guard) |
Controller: `Console.List/Get`.

### 10. `git` — R — "Show git sync status and pending changes."
| action | args | returns |
| --- | --- | --- |
| `status` | — | `{linked, state, ahead?, behind?, changes?, branch?, remote?}` |
| `changes` | — | `items:[{path,name,kind,change}]` |
| `session` | — | the open pull/conflict session: `items:[{id,name,kind,choices}]` |
Controller: `SyncManage.Status/Changes`, `SyncConflict.Session`.

### 11. `scripts` — R — "Check a script for errors or read the scripting reference."
| action | args | returns |
| --- | --- | --- |
| `check` | `code`, `phase` (`pre`\|`post`) | `{errors:[{line,column,message}]}` |
| `reference` | `topic?` | compact `pm.*` reference text |
| `snippets` | — | `items:[{id,name,code}]` |
Controller: `ScriptTools`.

### 12. `settings` — R — "Read app settings."
| action | args | returns |
| --- | --- | --- |
| `get` | — | `{script_timeout_ms, history_user_cap, history_ai_cap}` (AI access settings are not returned) |
Controller: `Settings.Get`.

---

## Write tools (10)

### 13. `workspace_edit` — W — "Create, open, switch or rename a workspace."
| action | args | notes |
| --- | --- | --- |
| `create` | `name`, `path` | `WorkspaceCreate.Handle` |
| `open` | `path` | `WorkspaceOpen.Handle` (an existing folder) |
| `switch` | `id` | `WorkspaceActivate.Handle`; the window follows (`WorkspacesChanged`) |
| `rename` | `id`, `name` | `WorkspaceRename.Handle` |
Notify: `WorkspacesChanged` (+ `WorkspaceTreeChanged` on switch).

### 14. `tree_edit` — D — "Create, update, rename, move or duplicate items, edit collection/folder settings, and manage OAuth tokens."
| action | args | notes |
| --- | --- | --- |
| `create` | `kind`, `parent_id?`, `name`; request: `method?`, `url?`, …; example: `request_id`, `from_last_response?:true` | `TreeManage.Create*`, `ExampleManage.Create/CreateFromResponse`. For `from_last_response`, the example's response body and headers are written **masked by the leak guard**, because an AI-made file is AI output (research R6). |
| `update` | `id` + any of `name, method, url, params, headers, body, auth, scripts, description` (request), or the example fields | `RequestSave` (merges onto the stored request), `ExampleManage.Save` |
| `rename` | `id`, `name` | `TreeManage.Rename` / `ExampleManage.Rename` |
| `move` | `id`, `parent_id`, `index?` | `TreeManage.Move` |
| `duplicate` | `id` | `TreeManage.Duplicate` / `ExampleManage.Duplicate` |
| `save_settings` | `id`, any of `auth, headers, scripts, variables, description` | `TreeSettings.SaveSettings` (collection variables: same secret rules as `env_edit.set_vars`) |
| `get_token` / `refresh_token` / `use_token` | `owner_id`, `use?` | `OAuth.GetNewToken/Refresh/Use` → metadata only |
Notify: `WorkspaceTreeChanged`; for an update also `RequestChangedOnDisk{id}`.

### 15. `env_edit` — D — "Create or update environments, switch the active one, and edit variables (secret values are write-only)."
| action | args | notes |
| --- | --- | --- |
| `create` | `name`, `color?`, `confirm_before_send?` | `EnvironmentManage.Create` |
| `update` | `id`, `name?`, `color?`, `confirm_before_send?` | `EnvironmentManage.Update` |
| `duplicate` | `id` | `EnvironmentManage.Duplicate` |
| `set_active` | `id` or `null` | `EnvironmentSetActive.Handle` |
| `add_var` | `scope`, `key`, `value?`, `secret?`, `description?` | `VariableTable.AddVariable` (secret → vault) |
| `set_vars` | `scope`, `vars:[{key, initial?, secret?, secret_value?, enabled?, description?}]` (full table) | `VariableTable.SaveTable`; a row going secret → plain → `MCP-02403`; `confirm_reveal` is never sent; an omitted `secret_value` keeps the vault value |
| `set_current` | `scope`, `key`, `value` | `VariableTable.SetCurrent` (secret rows → value goes to the vault) |
| `reset` / `persist` | `scope`, `key` | `VariableTable.Reset/Persist` (persist of a secret → existing refusal) |
Notify: `VariablesChanged`.

### 16. `docs_edit` — D — "Apply an edited OpenAPI document to a collection."
| action | args | notes |
| --- | --- | --- |
| `apply` | `collection_id`, `yaml`, `confirm_deletes?:false` | `TreeDocumentation.ApplyDocumentation`; deletes only when `confirm_deletes:true` |
Notify: `WorkspaceTreeChanged`.

### 17. `cookies_edit` — D — "Add or edit cookies and the cookie allowlist."
`upsert {domain,name,value,path?,expires?,secure?,http_only?}` → `Cookie.Upsert`; `allowlist {domains}` →
`Cookie.SaveAllowlist`. Notify: `CookiesChanged`.

### 18. `git_edit` — D — "Link, configure, commit, pull and resolve git sync for the workspace."
| action | args | notes |
| --- | --- | --- |
| `link` | `url`, `branch`, `subfolder?`, `account_id`, `use_repo_workspace?` | `SyncLink.Link` |
| `clone` | `url`, `branch`, `subfolder?`, `account_id`, `path` | `SyncLink.CloneAsWorkspace` |
| `settings` | `auto_pull?`, `auto_push?`, … | `SyncManage.SaveSettings` |
| `commit` | `message`, `paths?` (default: all) | `SyncTransfer.Commit` |
| `pull` | — | `SyncTransfer.Pull` (may open a conflict session) |
| `resolve` | `items:[{id, choice}]` | `SyncConflict.Resolve` |
| `finish_pull` / `cancel_pull` | — | `SyncConflict.FinishPull/CancelPull` |
Accounts are referenced by id from `git status`/link errors; account management is not exposed (FR-015). Notify:
`GitStatusChanged` (+ `WorkspaceTreeChanged` after a pull).

### 19. `settings_edit` — W — "Change app settings (script timeout, history caps)."
`save {script_timeout_ms?, history_user_cap?, history_ai_cap?}`. Merges onto the current settings, then
`Settings.Save`. Any `ai_access` key → `MCP-02403` (FR-006).

### 20. `import` — D — "Import Postman, Bruno or Xenon files from paths into the workspace."
| action | args | notes |
| --- | --- | --- |
| `detect` | `paths` | `Import.Detect` |
| `clashes` | `paths` | `Import.Clashes` |
| `run` | `paths`, `on_clash?` (`keep_both`\|`replace`\|`skip`), `passphrase?` (encrypted Xenon exports) | `Import.Import` |
| `discard` | — | `Import.Discard` |
Notify: `WorkspaceTreeChanged`, `VariablesChanged`.

### 21. `export` — D — "Export an item, environment or the workspace to a file path (never with secrets), or save a response body."
| action | args | notes |
| --- | --- | --- |
| `options` | `target` | `Export.Options` |
| `run` | `target`, `format` (`xenon`\|`postman`), `path`, `current_as_initial?` | `Export.Export` with `IncludeSecrets=false`; `include_secrets:true` → `MCP-02403` |
| `response_body` | `path` | For this session's last send. Text bodies are written **masked by the leak guard** (research R6). Binary bodies are written as received only when no known secret form appears in them; otherwise `MCP-02403`. |

### 22. `send` — D, open-world — "Send a saved or ad-hoc request with the active environment and return the response."
| args | notes |
| --- | --- |
| `request_id` **or** `request {method,url,params?,headers?,body?,auth?,scripts?}` | saved → `RequestGet` then `RequestSend`; ad hoc → `RequestSend` (nothing saved) |
| `send_unresolved?` | → `acknowledge_missing_secrets` (FR-029) |
| `send_expired_token?` | → `send_with_expired_token` (FR-029) |

Returns `{status, status_text, ms, size, headers:[[k,v]…], body?, truncated?, binary?, content_type, tests?:[{name,passed,error?}], error?}`. For text bodies, `size` and `truncated` refer to the **masked** text, so `response body` paging lines up. For binary bodies, `size` is the raw byte count.
Origin `ai` + client. Tab `mcp:<session-id>`. On `SND-05409` it goes through the approval flow (research R8) and can
return `MCP-03403` or `MCP-04408`. Notify: `HistoryChanged`, `CookiesChanged`.

---

## Destructive-only tools (2)

### 23. `delete` — D — "Delete an item, environment, cookie(s), workspace, OAuth token, git link, History or the console."
| target | args | controller |
| --- | --- | --- |
| `item` | `id` | `TreeManage.Delete` / `ExampleManage.Delete` |
| `environment` | `id` | `EnvironmentManage.Delete` |
| `cookie` | `domain`, `name`, `path?` | `Cookie.Delete` |
| `cookie_domain` | `domain` | `Cookie.DeleteDomain` |
| `cookies` | — | `Cookie.Clear` |
| `workspace` | `id`, `move_to_trash?:true` | `WorkspaceRemove.Handle` |
| `token` | `owner_id` | `OAuth.Clear` |
| `git_link` | — | `SyncLink.Unlink` |
| `history` | `source?` (`all`\|`user`\|`ai`) | `History.Clear` |
| `console` | — | `Console.Clear` |
Argument name: `target` (instead of `action`). Notify: matching event per target.

### 24. `git_push` — D, open-world — "Push committed changes to the linked remote."
No args. → `SyncTransfer.Push`. Notify: `GitStatusChanged`.

---

## Never exposed (FR-015)

Vault `Status`/`SetupPassphrase`/`Unlock`/`Reveal`/`CopySecret`/`Reset`; OAuth `Reveal`/`Copy`; all `GitAccount.*`;
`Dialog.*`; `Window.*`; `UIState.*`; `Snippet.Copy` (clipboard); `Request.ParseCurl/FormatBulk/ParseBulk/ApplyUrl/
ApplyParams/AutoHeaders` (editor helpers; the AI sends structured fields); `Import.PickSources`, `Export.PickTarget`,
`Export.ShowInFolder`; the `AIAccess` binding itself; `Settings` AI-access fields.
