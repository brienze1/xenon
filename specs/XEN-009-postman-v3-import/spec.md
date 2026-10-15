# Feature Specification: Import Postman v3 Collections, Environments and Git-Native Folders

**Feature Branch**: `xen-009`

**Created**: 2026-10-14

**Status**: Draft

**Input**: User description: "Import Postman v3 collections, environments, globals and git-native workspace folders. Xenon imports Postman v2.0/v2.1 JSON but rejects v3 with IMP-00415. v3 is what Postman's own tooling now produces (`postman collection get --json`, `postman workspace pull`). Import it with the same fidelity as v2.1, secrets to the vault, unsupported parts reported, clash handling as for other formats, through the UI dialog and the MCP `import` tool. Out of scope: exporting to v3 and fetching from Postman cloud. Source: https://github.com/brienze1/xenon/issues/4"

## Context

Builds on XEN-003 (Postman/Bruno import), XEN-006 (export) and XEN-008 (MCP `import` tool).
Postman's tooling now writes collections in its v3 format, in two shapes:

- **v3 JSON**: what `postman collection get <id> --json` prints. One document: a collection
  whose `items` nest folders (`$kind: collection`) and requests (`$kind: http-request`), with
  saved examples inline on each request.
- **Git-native YAML**: what `postman workspace pull` writes and what teams keep in git. A
  `postman/` folder with `collections/<collection>/` trees (one `*.request.yaml` per request, a
  `.resources/definition.yaml` per collection or folder, examples as `*.example.yaml` under
  `.resources/<request>.resources/examples/`), `environments/*.environment.yaml` and
  `globals/workspace.globals.yaml`.

Today a user who works with the Postman CLI must re-export every collection as v2.1 by hand to
move it into Xenon. This feature removes that step.

The v3 shapes in this spec were verified on 2026-10-14 against Postman CLI 1.73.0 output: real
collections fetched with `postman collection get --json`, collections converted with
`postman collection migrate`, and environments and globals written with
`postman environment`/`postman globals`.

## Clarifications

### Session 2026-10-14

- Q: A collection/folder `auth` list with more than one entry (Postman multi-auth) — what does
  import do? → A: Import the first entry and warn about the rest; no multi-auth in Xenon and no
  follow-up task. (The rule format is undocumented, and none of the 12 real collections surveyed
  had more than one entry.)
- Settled from Postman CLI 1.73.0 output rather than asked (2026-10-14): CLI JSON masks secret
  values as `********`; vault-linked variables carry `secret: true` + `source` and no value;
  globals live at `postman/globals/workspace.globals.yaml`; environments are
  `postman/environments/<name>.environment.yaml`; v3 JSON examples are inline, and YAML examples
  are files under `.resources/<request>.resources/examples/`; headers and credentials appear as
  both lists and maps.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Import a v3 collection JSON (Priority: P1)

Luis runs `postman collection get <id> --json > billing.json` (or an AI client does it for him
through the MCP `import` tool) and imports the file. Xenon recognises it as a Postman v3
collection and imports the same tree, requests, examples, scripts, auth, variables and
descriptions that the v2.1 export of the same collection would give him.

**Why this priority**: This is the smallest path from "signed in with the Postman CLI" to "the
collection is in Xenon", and it is the only shape an AI client can produce without writing files
into a git folder.

**Independent Test**: Import a v3 JSON collection and its v2.1 equivalent into two empty
workspaces; the two collection folders must be equal apart from ids.

**Acceptance Scenarios**:

1. **Given** a v3 collection JSON with nested folders, **When** Luis imports it as a new
   collection, **Then** the detect step names it "Postman collection v3", and the sidebar shows the
   same folders and requests in the same order as in the source `items` arrays.
2. **Given** requests that use every v3 body type (`json`, `text`, `xml`, `html`, `javascript`,
   `formdata` with text and file rows, `urlencoded`, `graphql`, `file`, `none`), **When**
   imported, **Then** each maps to the matching Xenon body, disabled rows stay disabled, and file
   references that do not exist on this machine are marked "re-pick needed" and listed in the
   summary.
3. **Given** auth on the collection, a folder and a request, and requests with no `auth` (inherit)
   or `auth.type: noauth`, **When** imported, **Then** each level keeps its auth, a missing auth
   becomes Inherit and `noauth` becomes No auth.
4. **Given** `beforeRequest`/`afterResponse` scripts at collection, folder and request level,
   **When** imported, **Then** they become the matching pre-request and post-response scripts at
   the same level.
5. **Given** a request with inline `examples`, **When** imported, **Then** each becomes a saved
   example on that request with its name, request (URL, method, headers, query params, body) and
   response (status code, status text, headers, body).
6. **Given** a v3 collection JSON pasted as text, **When** detected, **Then** it is treated exactly
   like the same JSON from a file.

---

### User Story 2 - Import a v3 git-native collection folder (Priority: P1)

Luis's team keeps its Postman collections in git as `postman/collections/<name>/`. He picks one
collection folder (or a zip of it) and gets that collection in Xenon.

**Why this priority**: Teams that already moved to Postman's git-native mode have no v2.1 file to
hand at all.

**Independent Test**: Import a collection folder produced by `postman collection migrate` from a
v2.1 fixture, and compare it with the import of that v2.1 fixture; the collection folders must be
equal apart from ids.

**Acceptance Scenarios**:

1. **Given** a collection folder with `.resources/definition.yaml`, subfolders with their own
   definitions, and `*.request.yaml` files, **When** imported, **Then** the collection name comes
   from the root definition's `name` (else the folder name), each folder's name from its
   definition's `name` (else the folder name), each request's name from its `name` field (else the
   file name without `.request.yaml`), and siblings are ordered by their `order` value.
2. **Given** a request whose examples live under
   `.resources/<request>.resources/examples/*.example.yaml`, **When** imported, **Then** each file
   becomes a saved example on that request, ordered by `order`.
3. **Given** headers, auth credentials or example headers written as a YAML map (`Content-Type:
   application/json`) instead of a list of `{key, value}` rows, **When** imported, **Then** both
   shapes give the same result.
4. **Given** a single `*.request.yaml` file, **When** imported into an existing collection or
   folder, **Then** that one request is added there with its examples, if its examples folder is
   next to it; importing it as a new top-level item is refused with the existing "A request must be
   imported into a collection or folder" error.

---

### User Story 3 - Import v3 environments and globals (Priority: P1)

Luis imports `postman/environments/staging.environment.yaml`, or the JSON that
`postman environment get <id> --json` prints, and gets a Xenon environment. Importing
`postman/globals/workspace.globals.yaml` (or `postman globals get --json` output) merges into
the workspace globals.

**Why this priority**: Requests are useless without the variables they reference, and secrets
must land safely.

**Independent Test**: Import an environment with plain, disabled and secret variables; check the
environment file holds no secret value and the vault holds it.

**Acceptance Scenarios**:

1. **Given** a v3 environment with plain variables, **When** imported, **Then** it becomes a
   Xenon environment named after its `name` (else the file name without `.environment.yaml`),
   with each variable's value as its initial value and `enabled: false` rows kept disabled.
2. **Given** a variable with `type: secret` and a value, **When** imported, **Then** the
   variable is a Xenon secret, its value is stored in the workspace vault, and the value appears
   in no workspace file, log or console output.
3. **Given** the CLI's JSON output, where secret values are masked as `********`, **When**
   imported, **Then** the variable is a secret with no stored value (shown as missing), and the
   mask is never stored as a value.
4. **Given** a globals file, **When** imported, **Then** its variables merge into the target
   workspace's globals under the same clash rules as v2.1 globals.

---

### User Story 4 - Import a whole `postman workspace pull` folder (Priority: P2)

Luis points Xenon at the `postman/` folder (or the repository folder that contains it, or a zip
of either) and gets every collection, environment and the globals in one import.

**Why this priority**: Moving a whole workspace in one step is the real goal, but each part
already works on its own through stories 1–3.

**Independent Test**: Import a folder with two collections, two environments and a globals file;
the preview lists five items, and after import all five exist in the chosen workspaces.

**Acceptance Scenarios**:

1. **Given** a folder containing `postman/collections/` (or the `postman/` folder itself, whose
   top level has `collections/`), **When** detected, **Then** the preview names it a "Postman
   workspace" and lists one item per collection folder directly under `collections/`, one per
   environment file under `environments/`, and one globals item if `globals/` holds a globals
   file.
2. **Given** that preview, **When** Luis imports, **Then** each item lands in the workspace
   chosen for it (the assignment works as for a Postman data dump), and clashes are resolved per
   item with keep both / replace / skip.
3. **Given** a pulled `postman/` folder that also holds parts Xenon does not import (e.g.
   `specs/`, `flows/`, `mocks/`), **When** imported, **Then** each part is reported in the summary
   as skipped with its file count, as a data dump's skipped parts are.
4. **Given** the repository folder that contains `postman/` (and `.postman/`, source code and so
   on), **When** imported, **Then** only the `postman/` folder is read. Nothing outside it is
   imported, reported or size-checked.

---

### User Story 5 - Same behaviour through the MCP `import` tool (Priority: P2)

An AI client that just ran `postman collection get --json` or `postman workspace pull` calls
Xenon's MCP `import` tool with the path, and gets the same detection, clashes, result and
warnings as the import dialog.

**Why this priority**: The issue's motivating flow is "an AI client using the Postman CLI
provides the files"; the dialog and the tool share one import service, so this needs only proof,
not new behaviour.

**Independent Test**: Run `detect` and `run` on the v3 fixtures through the MCP `import` tool and
compare with the dialog's results.

**Acceptance Scenarios**:

1. **Given** a v3 collection JSON path, **When** the MCP `import` tool runs `detect`, **Then** it
   reports kind "postman_collection", version "3" and the same items and warnings as the dialog.
2. **Given** a workspace-pull folder path, **When** the tool runs `run`, **Then** every item is
   imported with the tool's default clash handling.

---

### Edge Cases

- **Unsupported request protocols**: a `graphql-request` imports as an HTTP POST request with a
  GraphQL body (query and variables) and its scripts; `grpc-request`, `websocket-request`,
  `socket.io-request`, `mqtt-request`, `mcp-request` and `llm-request` are skipped, and each is
  listed in the summary as "`<name>` is a `<kind>` request, which Xenon does not support; it was
  skipped" (the same wording as the Bruno import).
- **Unsupported auth types** (e.g. `hawk`, `ntlm`, `edgegrid`, `akamai`, `asap`, `jwt`): kept for
  export and the request is sent without auth, with the same summary warning as v2.1.
- **Collection/folder auth given as a list** (Postman's multi-auth shape
  `[{id, name, type, credentials, rules?}]`): a one-entry list (what the Postman CLI writes for
  every collection/folder auth) imports as that auth. A list with more than one entry imports the
  first entry and the summary warns that the other entries (named) were not imported; Xenon
  has one auth per level. `rules` are not evaluated.
- **Collection, folder and request variables**: collection variables become Xenon collection
  variables; folder `variables` in a folder definition and any request-level variables are kept
  for export with the same "Xenon has no folder/request variables" warnings as v2.1.
- **Secret collection variables** (`secret: true`): imported as Xenon secret collection
  variables; the value goes to the workspace vault and the collection file holds only the name and
  `secret: true` (constitution II, "Secrets Never Leak"). An empty value gives a secret with no
  stored value.
- **Missing or duplicate `order`**: siblings with an `order` come first in ascending order;
  siblings without one follow; ties and the unordered group are sorted by file or folder name,
  case-insensitively, so the result does not depend on the operating system's directory order. In
  v3 JSON, the `items` array order is used as is.
- **A v3 file that does not parse, or an item with an unknown `$kind`** (inside a v3 folder or
  JSON, or a picked file named `*.request.yaml`): the import is refused with an error that names
  the file and line and says why ("`<path>` is not a valid Postman v3 file: …"). The reason never
  quotes file content, because the file may hold a secret value. An unknown `$kind` that ends in
  `-request` (a protocol Postman may add later) is not an error: it is skipped with the
  unsupported-request warning, like gRPC. Any other YAML file that does not parse cannot be
  recognised as v3 and keeps the existing "not a Xenon, Postman, Bruno or cURL export" error.
- **One broken file in a workspace folder**: the whole source is refused at detect with that
  file's error, and nothing is written. This matches a data dump with a broken collection today.
- **A `postman/` folder with nothing importable** (no files under `collections/`,
  `environments/` or `globals/`): it is not a Postman workspace and gets the existing "not a
  Xenon, Postman, Bruno or cURL export" error.
- **Sibling items with the same name** in a v3 collection: handled exactly as for a v2.1 import of
  the same tree (the existing file-name de-duplication).
- **A `.resources/` folder or a `definition.yaml` with no requests next to it**: an empty folder
  or collection is still imported, as with v2.1 empty folders.
- **Settings Xenon has no field for** (`maxRedirects`, `protocolVersion`,
  `disabledSystemHeaders`): kept for export, as v2.1 keeps the same keys in
  `protocolProfileBehavior`; `followRedirects` and `strictSSL` map to the request's
  follow-redirects and verify-SSL settings.
- **Secret flags**: an environment or globals variable is a secret when it has `type: secret`
  or `secret: true` (the CLI writes both on vault-linked variables).
- **Variables linked to a Postman vault secret** (`source: {provider: postman, postman:
  {secretId, vaultId, …}}`, no `value`): Xenon cannot reach Postman's vault, so the variable
  imports as a secret with no stored value, its `source` is kept for export, and the summary
  warns once per variable that the value must be entered in Xenon. The secret and vault ids are
  not secret values and may appear in the warning.
- **Pasted text**: only v3 JSON (collection, environment, globals) is detected from pasted text,
  as pasted v2.1 JSON is. Pasted YAML gets the existing "not a Xenon, Postman, Bruno or cURL
  export" error, because a git-native request is meaningless without its folder.
- **A v3 JSON environment or globals document has no scope marker**: it is a globals document
  when its file name ends with `.globals.yaml`, `.globals.yml` or `.globals.json`, when it sits in
  a `globals/` folder, when it has a top-level `workspace` field (what the CLI prints for cloud
  globals), or when its top-level `id` ends with `.globals.yaml` (what the CLI prints for local
  globals); otherwise it is an environment.
- **A folder that mixes Bruno and v3 files**, or a v3 folder that also holds an `archive.json`:
  the existing marker files (Postman dump, Bruno) keep their precedence, so detection results for
  every existing format do not change.
- **Large collections**: a v3 import of 1,000 requests behaves like the existing v2.1 large
  import (progress, cancel before writing).

## Requirements *(mandatory)*

### Functional Requirements

**Detection**

- **FR-001**: The import MUST stop refusing v3 sources with IMP-00415 ("Postman v3 YAML is not
  supported yet…") and instead detect them.
- **FR-002**: A JSON document whose top level has `$kind: "collection"` and an `items` array, and
  no v2 `info.schema`, MUST be detected as a Postman collection, version 3.
- **FR-003**: A folder (or a zip that unpacks to one) MUST be detected as a Postman collection,
  version 3, when it holds `*.request.yaml` files or a `.resources/definition.yaml` with
  `$kind: collection`, and it is not a Postman workspace folder (FR-005).
- **FR-004**: A single `*.request.yaml` file MUST be detected as a Postman request, version 3, and
  imported only into an existing collection or folder.
- **FR-005**: A folder MUST be detected as a Postman workspace, version 3, when its top level has
  `postman/collections/`, `postman/environments/` or `postman/globals/`; when its top level has
  `collections/` with v3 collection folders under it, `environments/` with `*.environment.yaml`
  files, or `globals/` with a `*.globals.yaml` file; or when the folder itself is named
  `collections` and its subfolders are v3 collection folders.
- **FR-006**: A YAML or JSON document with a top-level `name` and a `values` array of
  `{key, value, …}` rows and no `_postman_variable_scope` MUST be detected as a Postman environment
  or Postman globals, version 3, split by the rule in Edge Cases.
- **FR-007**: Detection of every format that is detected today MUST be unchanged; the marker
  files of a Postman data dump (`archive.json`) and of Bruno collections keep precedence over the
  v3 rules.

**Mapping (same fidelity as v2.1)**

- **FR-008**: Folders, requests and examples MUST be imported with their nesting and the order
  defined in Edge Cases; every imported item gets a new Xenon id and the source `id`, where
  present, is kept for export.
- **FR-009**: Requests MUST keep method, URL, query params (with disabled flags and
  descriptions), path variables (with descriptions), headers (with disabled flags and
  descriptions), body and description.
- **FR-010**: Bodies MUST map as follows: `json`, `text`, `xml`, `html`, `javascript` → raw body
  with that language; `formdata` → form-data with text and file rows, disabled flags and content
  types; `urlencoded` → urlencoded rows with disabled flags; `graphql` → GraphQL body with query
  and variables; `file` → binary body; `none` or no body → no body. File paths that do not exist
  on this machine are marked "re-pick needed" and listed in the summary, as for v2.1.
- **FR-011**: Auth MUST be imported at collection, folder and request level with the same type
  mapping as v2.1 (no auth, bearer, basic, API key, digest, OAuth 2.0, AWS Signature v4;
  others kept for export with the unsupported-auth warning). A missing `auth` MUST become Inherit
  and `type: noauth` MUST become No auth. Credentials MUST be accepted both as a list of
  `{key, value}` rows and as a map. An auth given as a list MUST import its first entry, and
  every further entry MUST be named in a summary warning (Clarifications).
- **FR-012**: Scripts MUST be imported at collection, folder and request level:
  `beforeRequest`/`http:beforeRequest` → pre-request, `afterResponse`/`http:afterResponse` →
  post-response. On a `graphql-request`, `beforeQuery` → pre-request and `afterResponse` →
  post-response. Collection or folder scripts of other types (`graphql:*`, `grpc:*`) only run
  for those protocols in Postman, so they are not imported as Xenon scripts. They are kept for
  export and reported in the summary.
- **FR-013**: Collection variables MUST become Xenon collection variables with their enabled
  flags and descriptions. Folder and request variables MUST be kept for export and reported with
  the v2.1 warnings.
- **FR-014**: Descriptions MUST be imported for collections, folders and requests.
- **FR-015**: Request settings `followRedirects` and `strictSSL` MUST map to the request's
  settings; other settings are kept for export (FR-018).
- **FR-016**: Environments and globals MUST import with each variable's value as initial value,
  `enabled` as the enabled flag, and `type: secret` or `secret: true` as a secret.
- **FR-017**: Anything Xenon cannot represent MUST appear in the import summary as a warning
  with the item and path it belongs to, and is never dropped silently. Cases the v2.1 or Bruno
  import already reports use the same warning code and wording. The v3-only cases (extra
  multi-auth entries, vault-linked secrets, non-HTTP-script types) get their own messages, and the
  plan's contract fixes their exact text.
- **FR-018**: The fields a v3 source carries that Xenon has no place for MUST be kept so that a
  later Postman export (v2.1) of the imported collection includes them, as v2.1 import keeps
  them.

**Secrets**

- **FR-019**: Secret environment, globals and collection-variable values MUST be stored in the
  workspace vault and
  MUST NOT be written to workspace files, logs, console output, or the import preview.
- **FR-020**: A secret whose value is the CLI mask `********`, or that is linked to a Postman
  vault secret (`source`), MUST be imported as a secret with no stored value; a vault link is
  kept for export and reported in the summary, one warning per variable.

**Workspace folders**

- **FR-021**: A Postman workspace import MUST list each collection, each environment and the
  globals as separate items, each with its own target workspace, using the same assignment step as
  a Postman data dump.
- **FR-022**: Parts of a Postman workspace folder that are not collections, environments or
  globals MUST be skipped and reported per part with their file counts. When the picked folder
  contains `postman/`, only `postman/` is read and reported.

**Behaviour shared with every format**

- **FR-023**: Name clashes MUST be resolved with keep both / replace / skip exactly as for other
  formats.
- **FR-024**: Writing MUST stay atomic per imported item, show progress and be cancellable before
  writing, as for other formats.
- **FR-025**: The import dialog and the MCP `import` tool MUST give the same detection, preview,
  warnings and result for v3 sources.
- **FR-026**: The PRD MUST list Postman v3 (JSON, git-native folders, environments, globals,
  workspace folders) among the supported import formats and drop the "Postman v3 YAML is not in
  the MVP" note.

### Key Entities

- **Postman v3 collection**: a tree of folders and requests with collection-level auth, scripts,
  variables and description; arrives as one JSON document or as a folder of YAML files.
- **Postman v3 request**: method, URL, query params, path variables, headers, body, auth,
  scripts, settings, description and saved examples.
- **Postman v3 example**: a saved request/response pair attached to a request.
- **Postman v3 environment / globals**: a named list of variables (key, value, enabled, type).
- **Postman workspace folder**: what `postman workspace pull` writes; holds collections,
  environments, globals and other parts Xenon does not import.
- **Import summary warning**: an existing entity; v3 imports produce the same kinds.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a fixture collection that covers every field in FR-008–FR-015, importing its
  v3 JSON, its v3 folder and its v2.1 JSON yields three collections that are equal apart from ids.
- **SC-002**: Importing a workspace folder with two collections, two environments and globals
  takes one import, and all five items exist afterwards.
- **SC-003**: In every v3 import scenario, no secret value appears in any workspace file, log or
  console output (0 occurrences).
- **SC-004**: Every v3 source part Xenon cannot represent produces exactly one summary warning;
  0 parts are dropped without one.
- **SC-005**: 100% of the import scenarios that pass before this feature still pass after it.
- **SC-006**: A 1,000-request v3 collection imports within the same time bound the existing
  1,000-request v2.1 scenario uses.

## Assumptions

- Exporting to v3 stays out of scope; a collection imported from v3 exports as v2.1 or Xenon.
- Xenon does not talk to Postman's cloud; the user or an AI client provides the files.
- Only the HTTP part of v3 is imported; GraphQL requests become HTTP requests with a GraphQL body;
  other protocols are skipped with a warning, as the Bruno import skips them.
- The v3 shapes are the ones Postman CLI 1.73.0 writes and prints (see Context); fields this spec
  does not name are kept for export (FR-018).
- Test fixtures are generated from v2.1 fixtures with `postman collection migrate` and hold no
  real secrets.
