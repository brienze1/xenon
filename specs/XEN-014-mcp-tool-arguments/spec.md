# Feature Specification: MCP Tools Declare Their Arguments; Clearer Workspace and Import Errors

**Feature Branch**: `xen-014`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "XEN-014 — AI clients connected to Xenon's MCP server must be able to discover and correctly call every tool, and get actionable errors. (1) Every MCP tool's published input schema declares every argument its actions read, with the right type, a short description and which actions need it, so a client such as Claude Code can call `import` with `paths` as an array. The declared arguments come from the same argument definitions the handlers decode, so they cannot drift. (2) Validation failures returned over MCP name the failing argument and the broken rule instead of a bare `invalid request`. (3) Importing a folder of loose standalone export files no longer fails with the misleading 'is not a Xenon, Postman, Bruno or cURL export', and a batch that mixes source kinds reports a sensible top-level kind. Source: GitHub issue brienze1/xenon#16."

## Context

Found while importing a whole Postman workspace (22 collections, 6 environments, globals) into a new
workspace through Xenon's MCP server from Claude Code. Verified against the code at `06d0d82`:

- Xenon publishes 24 MCP tools. Each tool's published input schema lists **only** its action selector
  (an enum of action names; `delete` names its selector differently). None of the arguments the
  actions actually read (`name`, `path`, `id`, `paths`, `source_id`, `on_clash`, `passphrase`, …)
  are declared, so clients guess. Claude Code sends undeclared arguments as strings, so `import`
  (whose `paths` must be a list) cannot be called from it at all:
  `MCP-07400 Invalid arguments for import: json: cannot unmarshal string into Go struct field MCPImport.paths of type []string`.
- Each tool's arguments are read by one or more argument definitions, one per action or group of
  actions (for example `env_edit` reads a different set of arguments for each of its actions).
- Validation already knows which field failed and which rule it broke, and the Xenon window
  already receives that detail. But the MCP error only carries `{code, message}`, and the message
  is the fixed text `invalid request`. So `workspace_edit create` without `path` answers
  `{"code":"WKS-00400","message":"invalid request"}`.
- A folder passed to import is only recognised when it has a known layout: a git-native Postman v3
  workspace or collection folder, a Postman data dump (`archive.json`), a Bruno collection, or a
  Xenon workspace folder. Any other folder fails with `IMP-00415 <folder> is not a Xenon, Postman,
  Bruno or cURL export`, even when it holds valid export files that import fine one by one.
- A detected batch reports the kind and version of its **first** input as the top-level kind. In
  the reported batch the globals file came first, so the batch was reported as `postman_globals`.

## Clarifications

### Session 2026-10-15 (specification)

- Q: When a folder matches no known layout but holds standalone export files at its top level, what
  happens? → A: **Import the files.** Each top-level file is treated as if it had been passed on
  its own (one level deep, no recursion). Files that are not importable on their own are skipped
  with a warning. If no file is importable, the import fails with an error that lists the folder
  layouts Xenon accepts.
- Q: What top-level kind does a batch that mixes source kinds report? → A: A new kind **`mixed`**,
  shown in the window as "Multiple sources". Each item already carries its own kind. A batch whose
  inputs all share one kind (and version) keeps that kind.
- Q: How do validation errors look over MCP? → A: **Message plus details.** The MCP error gains a
  `details` list of `{field, problem}` written in plain English (`is required`, `must be an
  absolute path`), and the message becomes e.g. `invalid request: path is required`. The window's
  existing error details get the same readable wording.

### Session 2026-10-15 (clarify — settled from the code, no user question needed)

- Q: Do stdio clients need separate work to see the new schemas and errors? → A: No. The
  `xenon mcp` stdio bridge only pumps JSON-RPC between stdin/stdout and the running server
  (`stdio_bridge.feature`), so whatever `tools/list` and tool calls return over HTTP reaches stdio
  clients unchanged.
- Q: How is a failing nested field named? → A: By its path from the arguments root, built from the
  JSON names, e.g. `variables[2].name`. A top-level field is just its name (`path`). No window code
  reads validation attributes (it only reads the details of `SND-05409`, `SND-06409` and
  `SND-07409`, `DOC-*` and sync errors, which are not validation failures), so this changes nothing
  the window depends on.
- Q: Which rules need plain-English wording? → A: Every rule the request definitions use today:
  `required`, `required_if`, `required_unless`, `not_blank`, `oneof`, `max`, `min`, `gt`, `gte`,
  `email`, `absolute_path`, `has_id`, `variable_scope`, `import_input`, `safe_subfolder`. `dive` and
  `omitempty` are modifiers, not rules.
- Q: How does a detail with several messages (e.g. `missing_secrets`) become `{field, problem}`? →
  A: `problem` is its messages joined with `, `.
- Q: (implementation) The full schemas make `tools/list` 34,505 characters, over XEN-008's
  24,000-character budget. Trim or raise? → A: **Raise the budget to 36,000 characters** and keep
  the full nested schemas.
- Q: Can per-action requirements be expressed as conditional schema? → A: No. Claude's tool-use API
  rejects `oneOf`/`anyOf`/`allOf` at the top level of an input schema, so per-action requirements go
  in descriptions (FR-006).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - An AI client discovers every tool argument and calls `import` (Priority: P1)

A developer using Claude Code (or any MCP client) asks it to import exported Postman files into
the active Xenon workspace. The client reads the tool list, sees that `import` takes `paths` as a
list of strings together with `on_clash`, `passphrase` and `source_id`, and calls it correctly the
first time.

**Why this priority**: Without declared arguments, `import` cannot be called from Claude Code at
all, and every other tool depends on the client guessing argument names. This is the bug.

**Independent Test**: List the tools over MCP and check that `import` and `workspace_edit` declare
every argument with its type, description and requirement. Then call `import run` with an array
`paths` and see the collection land in the workspace.

**Acceptance Scenarios**:

1. **Given** the MCP server is running, **When** a client lists the tools, **Then** the `import`
   tool's input schema declares `action` (enum `detect`, `clashes`, `run`, `discard`, required),
   `paths` (array of strings), `source_id` (string), `on_clash` (string, enum `keep_both`,
   `replace`, `skip`) and `passphrase` (string), each with a description.
2. **Given** the MCP server is running, **When** a client lists the tools, **Then** the
   `workspace_edit` tool's input schema declares `action` (enum `create`, `open`, `switch`,
   `rename`, required), `id`, `name` and `path` (strings), each with a description that names the
   actions using it and the actions requiring it (e.g. `path`: "Required for create and open.").
3. **Given** an active workspace, **When** a client calls `import` with action `run` and `paths` as
   an array of one Postman v2.1 collection file, **Then** the call succeeds and the collection
   appears in the workspace.
4. **Given** any of the 24 tools, **When** a client lists the tools, **Then** every argument any of
   that tool's actions reads is declared in its input schema with the right type.

---

### User Story 2 - Validation errors over MCP say which argument is wrong and why (Priority: P2)

The client calls `workspace_edit create` with a `name` but no `path`. Instead of `invalid request`,
it gets an error that says `path is required`, fixes the call, and retries.

**Why this priority**: Field-level errors make a failing call fixable without reading Xenon's
source. They matter most once arguments are discoverable (Story 1), but stand on their own.

**Independent Test**: Call `workspace_edit create` without `path`, then with a relative `path`, and
read the error each time.

**Acceptance Scenarios**:

1. **Given** the MCP server is running, **When** a client calls `workspace_edit` with action
   `create`, a `name` and no `path`, **Then** the call fails with code `WKS-00400`, message
   `invalid request: path is required`, and `details` containing `{field: "path", problem: "is required"}`.
2. **Given** the MCP server is running, **When** a client calls `workspace_edit create` with a
   relative `path`, **Then** the call fails with code `WKS-00400` and `details` containing
   `{field: "path", problem: "must be an absolute path"}`.
3. **Given** the MCP server is running, **When** a client calls `workspace_edit switch` without
   `id`, **Then** the call fails with code `WKS-08400`, message `invalid request: id is required`
   and `details` containing `{field: "id", problem: "is required"}`, naming the argument the
   client sends, not an internal name.
4. **Given** a call that breaks rules on two fields (no `name`, no `path`), **When** it is
   validated, **Then** the message lists both (`invalid request: name is required; path is required`)
   and `details` holds one entry per field.
5. **Given** the Xenon window creates a workspace through its own controls with an invalid path,
   **When** the error reaches the window, **Then** it still carries one detail per failing field
   (`attribute`, `messages`), now with the readable wording, and its description is the same
   readable message.

---

### User Story 3 - A folder of loose export files imports, and a mixed batch says "mixed" (Priority: P3)

The developer exported each collection, environment and the globals from Postman into one plain
folder and passes that folder to import, from the window or over MCP. Xenon finds the standalone
files inside it, previews them all, and reports the batch as "Multiple sources".

**Why this priority**: Passing the files one by one already works, so this is a convenience plus
an honest error and label. It completes the issue but is not blocking.

**Independent Test**: Detect a folder holding a Postman collection, a v2.1 environment and a v2.1
globals file, then a folder holding only unrelated files, and read the preview or error.

**Acceptance Scenarios**:

1. **Given** a folder whose top level holds a Postman v2.1 collection, a v2.1 environment and a
   v2.1 globals file and nothing else, **When** it is detected for import, **Then** the preview
   lists 3 items (collection, environment, globals) with no warnings and the top-level kind
   `mixed`.
2. **Given** that folder plus a `notes.txt` file and a subfolder, **When** it is detected, **Then**
   the 3 export files are listed, and one source warning names what was skipped (`notes.txt`, the
   subfolder).
3. **Given** a folder whose top level holds only files that are not importable on their own,
   **When** it is detected, **Then** detection fails with code `IMP-55415` and a message that names
   the folder and lists the accepted layouts: a git-native Postman v3 workspace or collection
   folder, a Postman data dump (`archive.json`), a Bruno collection (`bruno.json` or
   `opencollection.yml`), a Xenon workspace folder, or a folder of standalone export files.
4. **Given** a folder whose top level holds two Postman v2.1 collections only, **When** it is
   detected, **Then** the top-level kind is `postman_collection` with version `2.1`.
5. **Given** the paths of a v2.1 globals file followed by a v2.1 collection file, **When** they are
   detected together, **Then** the top-level kind is `mixed` with an empty version.
6. **Given** a recognised folder layout (git-native Postman v3, data dump, Bruno, Xenon workspace),
   **When** it is detected, **Then** it is read exactly as before (no change in behaviour).
7. **Given** a mixed batch in the window's import dialog, **When** the preview shows, **Then** the
   source line reads "Multiple sources" and the item table lists every item.

### Edge Cases

- A tool with a single action (e.g. `settings`, `docs_edit`) keeps its action optional, as today,
  and still declares every other argument.
- A tool with no actions (`send`, `git_push`) declares its arguments with no action selector.
- `delete` keeps its own selector name; the selector is declared as today, and its other
  arguments are declared too.
- An argument read by several actions with different requirements (e.g. `workspace_edit` `name`:
  required for `create` and `rename`, unused by `open`/`switch`) appears once, and its description
  names each action's requirement.
- Nested arguments (objects and lists of objects, e.g. request bodies or variable rows) are
  declared with their nested fields and types, not as an untyped object.
- Secret-bearing arguments (`passphrase`, secret variable values) are declared like any other
  argument: the schema exposes names and types, never values.
- Hidden files (names starting with `.`) in a loose folder are ignored silently, not reported.
- A loose folder file that is read but is not importable on its own (malformed JSON, a Postman v1
  collection, an unrelated JSON or YAML file, a zip whose extraction is refused) is skipped. The
  warning names it with its reason, and the rest still import. If every file is skipped, detection
  fails with the `IMP-55415` layout error.
- The per-file safety limits are unchanged for every folder: a document file over the single-file
  limit (200 MB) or not UTF-8. The folder is read before its layout is known, so such a file still
  fails the whole detect with its existing error, exactly as it does for a recognised layout today.
- Symbolic links and other non-regular entries in a loose folder are ignored silently and never
  followed, as the existing folder walk already ignores them.
- A loose folder is expanded from the path as passed, even when Xenon looked into a `postman/`
  subfolder while trying the git-native layout.
- A `.zip` export at a loose folder's top level is treated as if its path had been passed on its own.
- A loose folder inside a batch of paths expands in place: its files join the batch in name order.
- Error details for errors that are not validation failures (e.g. OAuth or OpenAPI errors that
  already carry details) keep their existing attribute and messages unchanged.

## Requirements *(mandatory)*

### Functional Requirements

**Tool argument schema**

- **FR-001**: Every published MCP tool MUST declare, in its input schema, every argument that any
  of its actions reads, under the same name the action reads it by.
- **FR-002**: Each declared argument MUST state its type: string, integer, number, boolean, array
  (with its item type), or object (with its nested properties). An argument that only accepts a
  fixed set of values MUST declare that set as an enum.
- **FR-003**: Each declared argument MUST carry a one-line description, and that description MUST
  say which actions use it and which require it, unless the tool has at most one action.
- **FR-004**: The action selector MUST stay declared as today: an enum of the tool's actions,
  listed as required when the tool has more than one action.
- **FR-005**: The declared arguments MUST be derived from the same argument definitions the actions
  decode, so adding, renaming or retyping an argument changes the published schema with no second
  edit. A tool's schema MUST NOT declare an argument none of its actions reads.
- **FR-006**: The input schema MUST NOT use `oneOf`, `anyOf` or `allOf` at its top level; per-action
  requirements live in the descriptions (FR-003), not in conditional schema.
- **FR-007**: With `paths` declared as an array of strings, `import` MUST accept an array `paths` from
  any MCP client, over both the HTTP endpoint and the stdio bridge.

**Validation errors**

- **FR-008**: When an MCP call fails, the error MUST carry `code`, `message` and, when the failure has
  details, a `details` list of `{field, problem}` entries. `field` and the field names in the
  message MUST be the argument names the client sends (e.g. `id`), even when the action
  validates the value under another internal name (e.g. `local_id`). Only fields the client
  cannot send keep their internal name.
- **FR-009**: Validation failures MUST describe each broken rule in plain English, for every rule
  in use (`required`, `required_if`, `required_unless`, `not_blank`, `oneof`, `max`, `min`, `gt`,
  `gte`, `email`, `absolute_path`, `has_id`, `variable_scope`, `import_input`, `safe_subfolder`), e.g.
  `is required`, `must not be blank`, `must be an absolute path`, `must be at most 200 characters`,
  `must be one of keep_both, replace, skip`. An unknown rule MUST read `is not valid`, never the bare
  rule name. A nested field is named by its path from the arguments root (`variables[2].name`).
- **FR-010**: A validation failure's message MUST be `invalid request: ` followed by each failing
  field and its problem, joined by `; `, in field order.
- **FR-011**: The Xenon window's errors MUST keep their existing shape (`code`, `description`,
  `error_details` with `attribute` and `messages`). Validation details MUST use the FR-009 wording
  and the description MUST use the FR-010 message.
- **FR-012**: Errors that are not validation failures MUST keep their current messages, and their
  existing details MUST appear in the MCP `details` list with `field` = the attribute and `problem`
  = its messages joined with `, `.
  - The whole error text, details included, MUST stay subject to the existing AI-output leak guard.
  - Details MUST never carry a secret value. Today's detail sources carry ids, names, colours,
    fingerprints, line and column numbers, provider error codes, next-step hints, and secret
    *names* (never values). Validation details carry field names and rule parameters, never the
    submitted value.

**Import of folders and batches**

- **FR-013**: When a folder matches none of the recognised layouts, import MUST read each regular,
  non-hidden file at the top level of the folder **as passed** as if its path had been passed on
  its own. Every existing per-file and archive safety limit applies to it unchanged.
- **FR-014**: Top-level files that are not importable on their own, and subfolders, MUST be skipped,
  and the preview MUST carry one source warning that names each skipped entry (with the reason, for
  files that failed to read).
- **FR-015**: When no top-level file of such a folder is importable, detection MUST fail with code
  `IMP-55415` and a message that names the folder and lists every accepted folder layout.
- **FR-016**: Folders with a recognised layout MUST be detected and read exactly as before.
- **FR-017**: A detected batch whose inputs all share one kind and version MUST report that kind and
  version. Any other batch MUST report the kind `mixed` with an empty version.
- **FR-018**: The window MUST label the `mixed` kind "Multiple sources" and show its items in the
  regular item table.
- **FR-019**: FR-013 to FR-018 MUST apply equally to imports started from the window and over MCP.

**Quality gates**

- **FR-020**: BDD scenarios MUST cover: `tools/list` exposing typed arguments for `workspace_edit`
  and `import`; `import run` with an array `paths`; `workspace_edit create` without `path` (and with
  a relative path) returning field-level errors; loose-folder import (success, skipped entries, no
  importable file); and the `mixed` kind. Every existing scenario MUST stay green, and `make validate`
  MUST pass.

### Key Entities

- **Tool argument**: one input a tool's action reads, with a name, type (including the item type
  for lists and the nested fields for objects), an optional enum, a description, and the actions
  that use or require it.
- **Tool input schema**: what a client sees for a tool: the action selector plus every tool
  argument.
- **Error detail**: one failing field and the plain-English problem with it. It appears as
  `{field, problem}` over MCP and as `{attribute, messages}` in the window.
- **Import source kind**: the top-level label of a detected batch. It is the shared kind of all
  its inputs, or `mixed`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 24 of 24 tools declare every argument their actions read. An automated check fails
  when any decoded argument is missing from the published schema.
- **SC-002**: An AI client can import a Postman export through `import run` on its first call,
  with no hand-built requests, when it passes `paths` as declared.
- **SC-003**: 100% of validation failures returned over MCP name the failing argument and the rule
  in words. No validation error message is the bare `invalid request`.
- **SC-004**: A folder of the 29 standalone Postman export files from the reported case can be
  imported with one path, with the same items and zero warnings as passing the 29 files one by one.
- **SC-005**: A batch that mixes source kinds is never labelled with a single input's kind.
- **SC-006**: All pre-existing BDD scenarios and unit tests pass unchanged, except tests that
  assert the old wording "failed on the '<rule>' rule", the bare "invalid request", the old
  folder message "<folder> is not a Xenon, Postman, Bruno or cURL export" for a folder that
  matches no layout, or the first-input kind of a mixed batch. Those are updated to the new wording.

## Assumptions

- Clients (including Claude Code) send arguments in the type the schema declares. Xenon does not
  coerce strings into lists or numbers.
- Arguments a call sends that the tool does not declare are still ignored, as today; this feature
  does not start rejecting extra arguments.
- Error codes do not change; only messages and details become readable.
- Loose-folder expansion goes one level deep. Nested folders inside a loose folder are skipped, not
  searched.
- There is no cap on the number of files in a loose folder beyond the existing per-file and
  archive limits. This matches today's folder import, which has none either.
- Nested argument objects (e.g. a full `Request` with its auth, body and scripts) are declared with
  all their fields, so `tools/list` grows from about 5,000 to about 34,500 characters.
  - XEN-008 set a `tools/list` budget of 24,000 characters (about 6,000 tokens, its SC-005).
  - The user decided (2026-10-15, during implementation) to raise that budget to **36,000
    characters (about 9,000 tokens)** rather than trim the nested schemas.
  - `test/integration/features/mcp/connect.feature` @tool_list asserts the new budget.
- The `mixed` kind adds no new import behaviour. Every item imports exactly as it would on its own.
- The window's import dialog needs only the new label; no layout change.
- Out of scope: new MCP tools or actions, renaming existing arguments, rejecting unknown
  arguments, new error codes, and showing per-field details inline in window forms (the window
  shows the readable description, as it does today).
