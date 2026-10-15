# Data Model: XEN-014

## pkg/jsonschema

### Property (`pkg/jsonschema/property.go`)

| Field | Type | Meaning |
|---|---|---|
| `Types` | `[]string` | JSON types. One entry (`string`, `integer`, `number`, `boolean`, `array`, `object`), or several when the `schema` tag lists them (`integer\|string`). |
| `Items` | `*Property` | Item schema for `array`. |
| `Properties` | `map[string]Property` | Nested fields for `object` from a struct. |
| `AdditionalProperties` | `*Property` | Value schema for `object` from a map. |
| `Enum` | `[]string` | From the `enum:"a,b,c"` tag. |
| `Description` | `string` | From the `description:"…"` tag. |
| `RequiredFor` | `[]string` | From the `required_for:"a,b"` tag. `always` = required on every call. |
| `UsedBy` | `[]string` | From the `used_by:"a,b"` tag. It narrows the registration's actions when one DTO serves several actions but a field is read by only some of them. |
| `ValidatedAs` | `[]string` | From the `validated_as:"local_id"` tag: the top-level names this argument takes in the request DTOs the actions validate, when they differ from the argument name. |

- `Properties(t reflect.Type) (map[string]Property, error)` reads the exported fields of struct
  `t` by their `json` names.
  - It skips `json:"-"` and `schema:"-"` fields.
  - An anonymous embedded struct without a `json` name (e.g. `MCPSendTarget` inside `MCPSend` and
    `MCPSnippetInput`) is flattened: its fields become top-level properties, as `encoding/json`
    decodes them.
  - It returns an error for a `json.RawMessage` field (`[]byte` named `RawMessage`) without a
    `schema` tag, and for an unsupported kind (chan, func, interface without a `schema` tag).
- `Property.Map() map[string]any` renders standard JSON Schema: `type`, `items`, `properties`,
  `additionalProperties`, `enum`, `description`. A single type renders as a string and several
  as an array. `RequiredFor`, `UsedBy` and `ValidatedAs` are not rendered.

### Struct tags (on integration DTOs)

| Tag | Example | Rule |
|---|---|---|
| `description` | `description:"Absolute path of the workspace folder."` | Mandatory on every top-level argument field of a registered DTO. |
| `required_for` | `required_for:"create,open"` / `required_for:"always"` | Names actions of the same tool; `always` puts the field in top-level `required`. |
| `enum` | `enum:"keep_both,replace,skip"` | Only for closed sets that the handler or validator enforces (`oneof`, enum types). |
| `schema` | `schema:"integer\|string"` / `schema:"-"` | Mandatory on `json.RawMessage` fields; `-` omits the field (only `ai_access`). |
| `used_by` | `used_by:"discard"` | Must be a subset of the registration's actions (the completeness test enforces it). Without it, the field is used by every action of its registration. |
| `validated_as` | `validated_as:"local_id"` | Every request-DTO name whose validation failure must be reported under this argument's name. |
| `schema_for` | `schema_for:"docs_edit"` | Only on fields of a DTO shared by two tools: the field is declared only by the named tools. |

## entrypoint/dto

### MCPToolArguments (`dto/mcp_tool_arguments.go`, new)

| Field | Type | Meaning |
|---|---|---|
| `Actions` | `[]string` | The tool actions that decode this DTO. Empty for tools without actions. |
| `Shape` | `reflect.Type` | The DTO type. |

Constructor `NewMCPToolArguments[T any](actions ...string) MCPToolArguments`.

### MCPTool (`dto/mcp_tool.go`, changed)

- Adds ``Arguments []MCPToolArguments `json:"-"` ``.
- `NewMCPTool(tool entity.AITool, arguments ...MCPToolArguments)`.
- `InputSchema() map[string]any`:
  - `type: object`.
  - `properties`: the selector as today (enum of actions), plus each argument field once:
    - the union over registrations;
    - "Used by" (from `Actions`, when the tool has more than one action) and "Required for" (from
      `RequiredFor`, except `always`) appended to the description;
    - the first registration wins on a type conflict at runtime, and the unit test fails on any
      conflict.
  - `required`: the selector when `RequiresAction()`, plus every `always` field, sorted.
  - No `oneOf`, `anyOf` or `allOf`.
- `ArgumentOf(action, attribute string) string` maps a validation attribute back to the argument
  the client sent (the action applies `@action` qualifiers). When the attribute's first path segment is listed in some argument's
  `ValidatedAs`, that segment is replaced by the argument name (`local_id` → `id`,
  `apply_to_all` → `on_clash`, `text` → `paths`). Otherwise the attribute is returned unchanged.
- `ToEntity()` is unchanged.

### MCPToolError (`dto/mcp_tool_error.go`, changed) and MCPToolErrorDetail (new)

```text
MCPToolError       { code: string, message: string, details?: [MCPToolErrorDetail] }
MCPToolErrorDetail { field: string, problem: string }
```

`NewMCPToolError(failure errs.BaseError, tool MCPTool, action string)` fills `details` from
`failure.ToDto("").Error.ErrorDetails`: `field = tool.ArgumentOf(action, Attribute)` and
`problem = strings.Join(Messages, ", ")`. `details` is omitted when empty.
For a validation failure (description starting `invalid request: `), the message is rebuilt from
the mapped details, so it names the argument the client sent: `invalid request: id is required`.
Other messages are kept as they are.

### Argument DTO tagging (33 DTOs)

Every top-level field of every registered DTO gets `description`. Fields that some action
requires get `required_for`, and closed sets get `enum`. Examples:

| DTO.field | Tags |
|---|---|
| `MCPWorkspaceEdit.Id` | `description:"Workspace id, from workspace list." used_by:"switch,rename" required_for:"switch,rename" validated_as:"local_id"` |
| `MCPWorkspaceEdit.Name` | `description:"Workspace name, at most 200 characters." used_by:"create,rename" required_for:"create,rename"` |
| `MCPWorkspaceEdit.Path` | `description:"Absolute path of the workspace folder." used_by:"create,open" required_for:"create,open"` |
| `MCPImport.Paths` | `description:"Absolute paths of the files or folders to import." used_by:"detect,clashes,run" required_for:"detect,clashes,run" validated_as:"text"` |
| `MCPImport.SourceId` | `description:"Source id returned by detect; omitted means nothing to discard." used_by:"discard"` |
| `MCPImport.OnClash` | `description:"How to settle a name clash; default keep_both on run." used_by:"clashes,run" enum:"keep_both,replace,skip" validated_as:"apply_to_all,items[].resolution"` |
| `MCPImport.Passphrase` | `description:"Passphrase of an export with encrypted secrets." used_by:"clashes,run"` |
| `MCPTreeList.Depth` | `description:"Levels to list: an integer of 1 or more, or \"all\"; default 1." schema:"integer\|string"` |
| `MCPSettingsSave.AIAccess` | `schema:"-"` |

`required_for` is derived from the action's validator (`required`/`not_blank` on the converted
request DTO) or the controller's explicit rejection, so the declaration matches what the call
enforces.

## validator

`errs.ErrorDetails` keeps its shape (`attribute`, `messages`).

- `Attribute` is now the JSON path without the root struct name (`path`, `variables[2].name`).
- `Messages` is now `[problem]`, using the R7 wording.
- `BadRequestError` description: `invalid request: <attribute> <problem>[; …]`.

## domain

### importkind (`enums/importkind/value.go`)

| Constant | Value | Visible |
|---|---|---|
| `Folder` | `folder` | No. An input-level marker for a folder with no known layout; it never reaches a source after expansion. |
| `Mixed` | `mixed` | Yes. The top-level kind of a batch with differing kinds or versions. |

### ImportInput (changed)

- New field `Path string`: the path exactly as passed, set by `importSourcePersistence.readPath`
  for every path input. `Root` stays the possibly re-rooted read folder; for example
  `readImportFolder` moves into `postman/` for git-native workspaces.
- `IsLooseFolder() bool` returns `Kind == importkind.Folder`.

### ImportSkippedEntry (`entity/import_skipped_entry.go`, new)

| Field | Type | Meaning |
|---|---|---|
| `Name` | `string` | The entry's name inside the folder (`notes.txt`, `sub`). |
| `Reason` | `string` | `folder`, or the description of the read/detect error. |

### ImportWarning (changed)

`NewSkippedEntriesWarning(folder string, entries []ImportSkippedEntry) ImportWarning` returns
code `skipped_part`, severity `warning`, and the message
`Skipped in <folder>: <name> (<reason>), …`. Each reason drops the leading `<name>` the
detection error already starts with: `notes.txt is not a … export` becomes
`notes.txt (not a … export)`.

### ImportSource (changed)

`NewImportSource(id, inputs, tempDirs, now)`:
- `Kind`/`Version` are the shared kind and version of all inputs;
- otherwise `Kind = mixed` and `Version = ""`;
- no inputs → empty, as today.

## application

### Port `usecase.ReadImportFolderFiles` (new)

```go
type ReadImportFolderFiles interface {
    ReadFolderFiles(ctx context.Context, folder entity.ImportInput) ([]entity.ImportInput, []string, []entity.ImportSkippedEntry, error)
}
```

- Lists `folder.Path`, not `Root`. Returns the inputs of its top-level regular, non-hidden
  files, in name order, plus the temp dirs created for zip files and the skipped entries.
- Symbolic links and other non-regular entries are ignored silently and never followed.
- A file whose read fails (zip extraction refused, size or encoding limit) becomes a skipped
  entry with the error description. The size and encoding limits of documents were already
  enforced by the initial folder walk, so in practice this concerns `.zip` files and non-document
  files.
- The error is only for failures to list the folder itself: `IMP-08404` when it is gone, a
  file error otherwise.

### `usecase.RejectUnimportableFolder(name string) error` (new, `detect_import_format.go`)

`ClientError` `IMP-00415`, HTTP 415, with the R9 message.

### State flow of a detect call

```text
paths ─► Read (persistence) ─► inputs
inputs ─► DetectImportFormat each ─► Folder? ──yes──► ReadFolderFiles ─► DetectImportFormat each file
                                     │                                   ├─ ok → expanded inputs (in place)
                                     │                                   └─ fail → skipped entry
                                     │                   none ok ─► RejectUnimportableFolder (IMP-00415)
                                     └─no──► input
expanded inputs ─► readers.read ─► items (+ skipped_part warning) ─► NewImportSource (shared kind | mixed)
```

## Appendix A — Complete argument tag table (verified against the code at `06d0d82`)

The registrations below and their tag values are final. "—" means no tag.

- **Registration** is the `dto.NewMCPToolArguments[T](actions…)` call in the tool's `Tool()`.
- **Used by** is set as `used_by` only when it narrows the registration.
- **Validated as** feeds `ArgumentOf`. Each entry is a request-DTO path (from the request DTO root,
  `[]` matching any index), optionally limited to some actions as `path@action|action`.

Every top-level field also gets a `description` tag: one sentence about the argument. A field
whose requirement depends on more than the action (alternatives, kinds) states it in that
sentence. Those sentences are noted in the last column.

| Tool | Registration | Field | Type (schema) | Used by | Required for | Enum | Validated as | Sentence must say |
|---|---|---|---|---|---|---|---|---|
| console | `MCPConsoleList` (list) | type | string | — | — | `request,log,error` | `filter` | empty = all |
| console | `MCPConsoleList` (list) | query | string | — | — | — | — | |
| console | `MCPConsoleList` (list) | cursor | string | — | — | — | — | from the previous page |
| console | `MCPConsoleGet` (get) | id | string | — | get | — | — | |
| cookies | `MCPCookiesList` (list) | domain | string | — | — | — | — | |
| cookies | `MCPCookiesList` (list) | cursor | string | — | — | — | — | |
| cookies_edit | `MCPCookieUpsert` (upsert) | domain | string | — | upsert | — | — | |
| cookies_edit | `MCPCookieUpsert` (upsert) | raw | string | — | upsert | — | — | a Set-Cookie header value |
| cookies_edit | `MCPCookieAllowlist` (allowlist) | domains | array of string | — | — | — | — | replaces the allowlist |
| delete | `MCPDelete` (all 10 targets) | id | string | item, environment, workspace | item, environment, workspace | — | `local_id@workspace`, `example_id@item` | |
| delete | `MCPDelete` | move_to_trash | boolean | workspace | — | — | — | default true |
| delete | `MCPDelete` | domain | string | cookie, cookie_domain | cookie, cookie_domain | — | — | |
| delete | `MCPDelete` | name | string | cookie | cookie | — | — | |
| delete | `MCPDelete` | path | string | cookie, git_link | git_link | — | `target_folder@git_link` | absolute folder for git_link |
| delete | `MCPDelete` | owner_id | string | token | token | — | — | |
| delete | `MCPDelete` | source | string | history | — | `all,user,ai` | — | |
| docs | `MCPDocsInput` (get, plan) | collection_id | string | — | get, plan | — | `id` | |
| docs | `MCPDocsInput` (get, plan) | yaml | string | plan, apply (`used_by:"plan,apply"`) | — | — | — | |
| docs_edit | `MCPDocsInput` (apply) | collection_id | string | — | apply | — | `id` | |
| docs_edit | `MCPDocsInput` (apply) | yaml | string | — | — | — | — | |
| docs_edit | `MCPDocsInput` (apply) | confirm_deletes | boolean | — | — | — | — | `schema_for:"docs_edit"` (see note 1) |
| env | `MCPEnvScope` (get) | scope | string | — | get | — | — | `globals`, `env:<id>` or `collection:<id>` |
| env | `MCPEnvInspect` (inspect) | text | string | — | — | — | — | inspected when no request_id |
| env | `MCPEnvInspect` (inspect) | request_id | string | — | — | — | `id` | |
| env_edit | `MCPEnvEdit` (create, update, duplicate, reorder) | name | string | create, update | create | — | — | |
| env_edit | `MCPEnvEdit` | color | string | create, update | — | `none,green,cyan,magenta,yellow,red` | — | |
| env_edit | `MCPEnvEdit` | confirm_before_send | boolean | create, update | — | — | — | |
| env_edit | `MCPEnvEdit` | index | integer | reorder | reorder | — | — | 0-based |
| env_edit | `MCPEnvEdit` | id | string | update, duplicate, reorder | update, duplicate, reorder | — | — | |
| env_edit | `MCPEnvActivation` (set_active) | id | string | — | — | — | — | omit or empty to deactivate (merges with `MCPEnvEdit.id`, see note 2) |
| env_edit | `MCPEnvVariable` (add_var, set_current, reset, persist) | scope | string | — | add_var, set_current, reset, persist | — | — | |
| env_edit | `MCPEnvVariable` | key | string | — | add_var, set_current, reset, persist | — | — | |
| env_edit | `MCPEnvVariable` | value | string | add_var, set_current | — | — | — | |
| env_edit | `MCPEnvVariable` | description | string | add_var | — | — | — | |
| env_edit | `MCPEnvVariable` | secret | boolean | add_var | — | — | — | |
| env_edit | `MCPEnvVariables` (set_vars) | scope | string | — | set_vars | — | — | (merges with `MCPEnvVariable.scope`) |
| env_edit | `MCPEnvVariables` (set_vars) | vars | array of object (`MCPVariableRow`: key, initial, description, secret, enabled, secret_value) | — | — | — | — | replaces the table; secret_value only on secret rows |
| export | `MCPExport` (options, run) | target | object (`MCPExportTarget`: kind enum `workspace,collection,folder,request,environment,globals` with `validated_as:"item_kind"`, id with `validated_as:"item_id"`) | — | options, run | — | (nested) | id not needed for workspace or globals |
| export | `MCPExport` | format | string | run | run | `xenon,xenon_yaml,postman` | — | |
| export | `MCPExport` | path | string | run | run | — | `target_path` | absolute |
| export | `MCPExport` | current_as_initial | boolean | run | — | — | — | |
| export | `MCPExport` | include_secrets | boolean | run | — | — | — | refused over MCP: humans only |
| export | `MCPSaveResponseBody` (response_body) | path | string | — | response_body | — | — | absolute (merges with `MCPExport.path`) |
| git_edit | `MCPGitLink` (link, clone) | url | string | — | link, clone | — | `repository` | |
| git_edit | `MCPGitLink` | branch | string | — | link | — | — | |
| git_edit | `MCPGitLink` | subfolder | string | — | — | — | — | |
| git_edit | `MCPGitLink` | account_id | string | — | link, clone | — | — | |
| git_edit | `MCPGitLink` | path | string | — | — | — | `clone_path` | absolute clone folder |
| git_edit | `MCPGitLink` | name | string | clone | — | — | — | |
| git_edit | `MCPGitLink` | use_repo_workspace | boolean | link | — | — | `mode` | |
| git_edit | `MCPGitSettings` (settings) | auto_pull / auto_push | boolean | — | — | — | — | omitted keeps the stored value |
| git_edit | `MCPGitSettings` (settings) | account_id | string | — | — | — | — | (merges with `MCPGitLink.account_id`) |
| git_edit | `MCPGitCommit` (commit) | message | string | — | commit | — | — | |
| git_edit | `MCPGitCommit` (commit) | item_ids | array of string | — | — | — | — | omitted = every changed item |
| git_edit | `MCPGitResolve` (resolve) | items | array of object (`MCPGitResolution`: id `validated_as:"item_id"`, choice enum `mine,theirs,both,keep,delete` `validated_as:"resolution"`, field) | — | — | — | (nested) | |
| history | `MCPHistoryList` (list) | source | string | — | — | `all,user,ai` | — | |
| history | `MCPHistoryList` (list) | query / cursor | string | — | — | — | — | |
| history | `MCPHistoryGet` (get) | id | string | — | get | — | — | |
| import | `MCPImport` (detect, clashes, run, discard) | paths | array of string | detect, clashes, run | detect, clashes, run | — | `text` | |
| import | `MCPImport` | on_clash | string | clashes, run | — | `keep_both,replace,skip` | `apply_to_all`, `items[].resolution` | default keep_both on run |
| import | `MCPImport` | passphrase | string | clashes, run | — | — | — | |
| import | `MCPImport` | source_id | string | discard | — | — | — | |
| response | `MCPResponseBody` (body) | offset / length | integer | — | — | — | — | length capped at 65536 |
| scripts | `MCPScriptCheck` (check) | text | string | — | — | — | — | |
| send | `MCPSend` (no actions) | request | object (`Request`) | — | — | — | — | pass exactly one of request or request_id |
| send | `MCPSend` | request_id | string | — | — | — | `id` | |
| send | `MCPSend` | send_unresolved / send_expired_token | boolean | — | — | — | — | |
| settings_edit | `MCPSettingsSave` (save) | script_timeout_ms / history_user_cap / history_ai_cap | integer | — | — | — | — | |
| settings_edit | `MCPSettingsSave` | ai_access | — | — | — | — | — | `schema:"-"` (not declared) |
| snippet | `MCPSnippetInput` (generate) | request | object (`Request`) | — | — | — | — | pass exactly one of request or request_id |
| snippet | `MCPSnippetInput` | request_id | string | — | — | — | `id` | |
| snippet | `MCPSnippetInput` | language | string | — | generate | `curl,js_fetch,go_native,python_requests` | — | |
| snippet | `MCPSnippetInput` | include_secrets | boolean | — | — | — | — | refused over MCP: humans only |
| tree | `MCPTreeList` (list) | parent_id / cursor | string | — | — | — | — | |
| tree | `MCPTreeList` (list) | depth | `integer\|string` (`schema` tag) | — | — | — | — | an integer of 1 or more, or "all"; default 1 |
| tree | `MCPTreeSearch` (search) | query | string | — | search | — | — | |
| tree | `MCPTreeSearch` (search) | limit | integer | — | — | — | — | 0–100 |
| tree | `MCPTreeTokenOwner` (token) | owner_id | string | — | token | — | — | |
| tree | `MCPTreeItem` (get, settings) | id | string | — | get, settings | — | `example_id` | |
| tree_edit | `MCPTreeEdit` (all 9 actions) | kind | string | create | create | `collection,folder,request,example` | — | |
| tree_edit | `MCPTreeEdit` | name | string | create, update, rename | rename | — | — | required to create a collection, folder or request |
| tree_edit | `MCPTreeEdit` | parent_id | string | create, move | — | — | — | required to create a folder or request |
| tree_edit | `MCPTreeEdit` | id | string | update, rename, move, duplicate, save_settings | update, rename, move, duplicate, save_settings | — | `example_id` | |
| tree_edit | `MCPTreeEdit` | index | integer | move | — | — | — | |
| tree_edit | `MCPTreeEdit` | request_id | string | create | — | — | — | required to create an example |
| tree_edit | `MCPTreeEdit` | from_last_response | boolean | create | — | — | — | |
| tree_edit | `MCPTreeEdit` | method / url | string | create, update | — | — | — | |
| tree_edit | `MCPTreeEdit` | description | string | create, update, save_settings | — | — | — | |
| tree_edit | `MCPTreeEdit` | params / headers | array of object (`KeyValue`) | params: create, update; headers: create, update, save_settings | — | — | — | |
| tree_edit | `MCPTreeEdit` | body | object (`Body`) | create, update | — | — | `request.body`, `example.request.body` | |
| tree_edit | `MCPTreeEdit` | auth | object (`Auth`) | create, update, save_settings | — | — | `request.auth`, `example.request.auth` | |
| tree_edit | `MCPTreeEdit` | scripts | object (`Scripts`) | create, update, save_settings | — | — | `request.scripts`, `example.request.scripts` | |
| tree_edit | `MCPTreeEdit` | response | object (`ExampleResponseEdit`) | update | — | — | `example.response` | |
| tree_edit | `MCPTreeEdit` | variables | array of object (`MCPVariableRow`) | save_settings | — | — | — | collections only |
| tree_edit | `MCPTreeEdit` | owner_id | string | get_token, refresh_token, use_token | get_token, refresh_token, use_token | — | — | |
| tree_edit | `MCPTreeEdit` | use | string | use_token | use_token | `access_token,id_token` | `use_token` | |
| workspace_edit | `MCPWorkspaceEdit` (create, open, switch, rename) | id | string | switch, rename | switch, rename | — | `local_id` | |
| workspace_edit | `MCPWorkspaceEdit` | name | string | create, rename | create, rename | — | — | at most 200 characters |
| workspace_edit | `MCPWorkspaceEdit` | path | string | create, open | create, open | — | — | absolute |

Tools that read no arguments in some actions register nothing for those actions. That is the
case for `workspace list|active`, `git status|changes|session`, `settings get`,
`snippet languages`, `scripts reference|snippets`, `response last`, `cookies allowlist`,
`env list`, `git_edit pull|finish_pull|cancel_pull` and `git_push`. Their
selector is still declared.

**Note 1.** `docs` and `docs_edit` are separate tools that share `MCPDocsInput`.
- `docs` registers `NewMCPToolArguments[MCPDocsInput]("get", "plan")`. `confirm_deletes` is not
  read by `docs`, so it is tagged `schema_for:"docs_edit"`. That is the one tool-scope tag: a
  field with `schema_for` is declared only by the named tools.
- `yaml` gets `used_by:"plan,apply"`, and `collection_id` gets `required_for:"get,plan,apply"`.
  `used_by` and `required_for` are read against each registering tool's actions; an action that
  tool lacks is ignored. The completeness test fails only when a name matches no action of any
  tool that registers the DTO.

**Note 2.** When two registrations of one tool declare the same field (`env_edit.id`,
`env_edit.scope`, `export.path`, `git_edit.account_id`), the schema property is declared once:
- its type must agree, or the completeness test fails;
- its "Used by" is the union of their actions;
- its "Required for" is the union of their `required_for`;
- its description is the first registration's sentence.

**Note 3.** `validated_as` paths are matched against the validation attribute (the JSON path
from the request DTO root). The longest matching prefix wins and is replaced by the argument's
own path (`target.kind` for a nested field, `items[].id` inside arrays). Attributes that match
nothing are kept as they are; those are internal fields such as `local_id` or `tab_id` that the
client cannot send. A path qualified with `@action` applies only to that action.
