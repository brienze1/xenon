# Feature Specification: Import (Postman, Bruno, cURL) and Saved Examples (Stage 3)

**Feature Branch**: `xen-003`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 3 — Import Postman collections/environments/globals/data dumps, Bruno collections and cURL commands; saved examples as a Xenon feature; fidelity so Postman → Xenon → Postman loses nothing. Source: docs/prd/xenon-mvp-prd.md 'Import and export'; builds on XEN-001 and XEN-002."

## Context

Stage 3 of 7. The PRD wants real work in Xenon early so gaps show up fast: this stage brings
Luis's existing Postman (and older Bruno) collections and environments in, sorted into his four
company workspaces. Scripts are imported and stored now; they run from stage 4. Unsupported
auth types are kept for export and become usable in stage 5. Export (including Xenon YAML
import, which ships alongside it) is stage 6, and must round-trip what this stage imports.

## Clarifications

### Session 2026-10-09

- Q: Import scope? → A: Postman collection v2.0/v2.1 JSON, environments, globals and data dumps;
  Bruno `.bru` and OpenCollection YAML; cURL. Xenon YAML import is stage 6. Postman's v3 YAML
  (Native Git) is not in the MVP.
- Q: Postman data dumps carry no workspace membership — how to import? → A: One screen lists
  every collection and environment in the dump with a target workspace per item (new workspaces
  can be created inline).
- Q: Bruno scripts use `bru`/`req`/`res`, Xenon uses `pm.*` — how to import? → A: Translate the
  common calls to `pm.*`; flag anything untranslated.
- Q: Postman saved examples? → A: Make them a Xenon feature (sidebar, tab, Save response as
  example, edit) and round-trip them.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Import a Postman collection (Priority: P1)

Luis drops a Postman collection JSON (v2.0 or v2.1) on Xenon, picks the target workspace and
where to put it (as a new collection, or into an existing collection or folder), and gets every
folder and request with the same order, methods, URLs, params, headers, bodies, auth, scripts,
variables, descriptions and examples. A summary tells him what came in and anything that needs
attention.

**Why this priority**: Moving real collections over is what lets Luis start the switch; the PRD
says "real data in Xenon early".

**Independent Test**: Import a representative collection (nested folders, every body mode, every
auth type, scripts at all levels, collection variables, examples), then compare it field by field
with the source and read the summary.

**Acceptance Scenarios**:

1. **Given** a Postman v2.1 collection with nested folders, **When** Luis imports it into Acme as
   a new collection, **Then** the sidebar shows the same tree in the same order, and each request
   has the same method (including custom/WebDAV methods), URL, query params (with disabled rows
   and descriptions), path variables, headers, body and auth as in Postman.
2. **Given** requests with every body mode (raw with language, urlencoded, form-data text and
   file, file/binary, GraphQL, and a disabled body), **When** imported, **Then** each maps to the
   matching Xenon body; form-data/binary files whose paths do not exist on this machine are
   marked missing and listed in the summary.
3. **Given** auth on the collection, folders and requests — inherit (no auth key), `noauth`,
   bearer, basic, apikey, oauth2, awsv4, digest, plus types Xenon does not support (hawk, ntlm,
   edgegrid, oauth1, jwt, asap) — **When** imported, **Then** inheritance and explicit "no auth"
   are preserved, supported types map to Xenon's auth, and unsupported types are kept unchanged
   for export, shown as "unsupported auth type (kept)" and listed in the summary.
4. **Given** pre-request and test scripts on the collection, folders and requests, **When**
   imported, **Then** they become the matching pre-request / post-response scripts with the same
   text; disabled scripts stay disabled; scripts using `pm.require` or packages are listed in the
   summary.
5. **Given** collection variables (with types, disabled rows and descriptions), **When**
   imported, **Then** they become collection variables (initial values); Postman folder- or
   request-level variables, which Xenon has no place for, are kept for export and listed in the
   summary.
6. **Given** requests with saved examples, **When** imported, **Then** each example appears under
   its request (User Story 4).
7. **Given** Postman-only settings (`protocolProfileBehavior` such as follow redirects, strict
   SSL, URL encoding), **When** imported, **Then** the ones Xenon has (follow redirects, SSL
   verification, encode URL) map to request settings and all of them are kept for export.
8. **Given** the import finished, **Then** a summary shows counts (collections, folders,
   requests, examples, variables) and every warning with the item it concerns, and lets Luis
   open any listed item.

---

### User Story 2 - Import Postman environments and globals (Priority: P1)

Luis imports his Postman environment JSON files (and the globals export) into the right
workspace. Plain values become initial values; values Postman marks as secret become Xenon
secrets, stored in the vault — never in the files.

**Why this priority**: Requests without their environments cannot be sent; secrets must be safe
from the first import.

**Independent Test**: Import environments with plain, secret (with and without value) and
disabled variables plus a globals file; check files, vault and the missing markers.

**Acceptance Scenarios**:

1. **Given** a Postman environment with plain variables, **When** imported into Acme, **Then** an
   environment with the same name and variables (initial values, enabled state) exists.
2. **Given** variables of type `secret` with values, **When** imported, **Then** they become
   Xenon secret variables whose values are in the vault and in no workspace file; secret
   variables exported without a value become secrets marked "missing".
3. **Given** a Postman globals export, **When** imported, **Then** its variables are merged into
   the workspace globals following the clash rule (FR-018).
4. **Given** an environment name that already exists in the workspace, **Then** Luis chooses
   replace, keep both (imported one gets " (imported)") or skip.

---

### User Story 3 - Sort a whole Postman data dump into workspaces (Priority: P2)

Luis exports all his Postman data (a zip, or the unzipped folder) and imports it in one go. Xenon
lists every collection and environment found; for each, he picks the target workspace (or creates
a new one right there), then imports everything at once.

**Why this priority**: Saves importing dozens of files one by one across four companies; useful
once single imports work.

**Independent Test**: Import a dump with collections and environments from several companies,
assign them to three workspaces (one created inline), and verify each workspace got exactly its
items.

**Acceptance Scenarios**:

1. **Given** a Postman data dump (zip or folder), **When** Luis opens it in the import dialog,
   **Then** every collection and environment is listed with its name, kind and size, each with a
   target-workspace picker defaulting to the current workspace, and a checkbox to include it.
2. **Given** he chooses "New workspace…" in a picker, **Then** he can name it and pick a folder
   without leaving the screen, and it becomes available in all pickers.
3. **Given** assignments are made, **When** he imports, **Then** each item lands in its chosen
   workspace, clashes are resolved per FR-018, and one combined summary lists the results per
   workspace.
4. **Given** a dump that contains anything other than collections and environments, **Then**
   those parts are skipped and listed in the summary.

---

### User Story 4 - Saved examples (Priority: P2)

A request can have saved examples: a named pair of a request snapshot and a response (status,
headers, body). They appear under the request in the sidebar, open in a tab, can be created from
the response of the current request ("Save response as example"), and can be renamed,
duplicated, edited and deleted. They round-trip with Postman.

**Why this priority**: Luis relies on examples as documentation of what an endpoint returns;
explicitly requested.

**Independent Test**: Send a request, save the response as an example, edit it, restart, and
check it is listed, opens and persists; import a Postman request with examples and check they
appear.

**Acceptance Scenarios**:

1. **Given** a request with examples, **Then** the sidebar shows them nested under the request
   (expandable), each with its name and response status.
2. **Given** Luis clicks an example, **Then** it opens in a tab (preview/pin rules as for
   requests) showing the example's request (method, URL, params, headers, body) and its response
   (status code and text, headers, body with Pretty/Raw/Preview), all editable; changes are saved
   with Save.
3. **Given** a response in a request tab, **When** he chooses "Save response as example",
   **Then** he names it and an example is created from the request as sent (unresolved
   `{{variables}}`) and the response received.
4. **Given** an example, **When** he renames, duplicates or deletes it from the sidebar menu,
   **Then** it changes accordingly and is stored with its request.
5. **Given** an example's response body larger than 5 MB, **Then** it is stored and opens in Raw
   first, as for live responses.

---

### User Story 5 - Import Bruno collections (Priority: P2)

Luis's older collections are in Bruno, either as `.bru` files or Bruno's newer YAML format. He
imports a Bruno collection folder (or zip) and gets the same tree, requests, environments
(secret names only — Bruno keeps values elsewhere) and scripts translated to `pm.*` where
possible.

**Why this priority**: Covers the collections Luis already moved to Bruno; secondary to Postman.

**Independent Test**: Import a `.bru` collection and a YAML collection with folders, ordering,
all body modes, auth, request vars, assertions, scripts and environments; compare and read the
summary.

**Acceptance Scenarios**:

1. **Given** a Bruno collection folder in `.bru` format or YAML format, **When** imported,
   **Then** folders and requests appear in Bruno's order (its sequence numbers), with method,
   URL, query and path params, headers (disabled rows included), body, auth (including inherit),
   docs (as description) and settings.
2. **Given** Bruno environments, **Then** they become environments; plain vars become initial
   values; secret vars become Xenon secrets marked "missing" (Bruno does not store their values
   in the collection).
3. **Given** Bruno scripts (pre-request, post-response, tests) using `bru`, `req`, `res`,
   `test()` and `expect()`, **When** imported, **Then** common calls are rewritten to `pm.*`
   equivalents; any line that could not be translated is kept with a `// XENON: untranslated`
   comment and listed in the summary with its location.
4. **Given** Bruno request variables (`vars:pre-request`, `vars:post-response`) and assertions
   (`assert`), **Then** they are converted into equivalent pre-request / post-response script
   lines (variable sets and `pm.test` checks); anything not convertible is listed in the summary.
5. **Given** a collection mixing both formats or not recognised as Bruno, **Then** the import is
   refused with a clear message.

---

### User Story 6 - Import a cURL command (Priority: P2)

Luis copies a cURL command (from a terminal, browser DevTools, docs, or Windows cmd/PowerShell)
and pastes it into the URL bar or the import dialog. Xenon turns it into a request.

**Why this priority**: Very common daily shortcut; independent of the other imports.

**Independent Test**: Paste bash, cmd (`^`) and PowerShell (backtick) variants of the same command
and compare the resulting requests.

**Acceptance Scenarios**:

1. **Given** Luis pastes `curl -X POST https://api.test/x -H 'Content-Type: application/json'
   -d '{"a":1}'` into the URL bar of a tab, **Then** that tab's method, URL, headers and body are
   replaced by the parsed ones (raw JSON body), and he can undo the replacement.
2. **Given** the same command pasted into the import dialog, **Then** a new request is created
   in the chosen collection/folder (name derived from method and path, editable).
3. **Given** commands using `-u user:pass`, `-F` form fields and files, `--data-urlencode`, `-G`,
   `-b` cookies, `-A`, `-e`, `--json`, `-I`, `-T`, `--url`, `--compressed`, `-k`, `-L`, **Then**
   each maps as defined in FR-024 (supported flags) or is ignored silently (no-effect flags), and
   unknown flags are listed in a notice.
4. **Given** multi-line commands with `\` (bash), `^` (Windows cmd) or `` ` `` (PowerShell)
   continuations and their quoting, **Then** all three produce the same request.
5. **Given** pasted text in the URL bar that is not a cURL command, **Then** it is treated as a
   normal URL (no parsing attempt beyond detecting a leading `curl `).

---

### Edge Cases

- **Invalid or unsupported files** (malformed JSON, Postman v1, Postman v3 YAML, unknown
  formats): the import is refused before anything is written, naming the problem and, for v3
  YAML, saying it is not supported in this version.
- **Partial failure**: an import is all-or-nothing per target collection/environment — if writing
  fails midway, nothing from that item remains on disk, and the summary says which items failed.
- **Name clashes** (collection, environment, request inside an existing folder): replace, keep
  both (" (imported)" suffix) or skip, chosen once with "apply to all" for the rest.
- **Very large collections** (thousands of requests, multi-MB example bodies): import shows
  progress, can be cancelled before writing starts, and stays responsive.
- **Duplicate Postman ids** across items: each imported item gets a fresh Xenon id; the Postman
  ids are kept for export.
- **Header as raw string, `request` as plain URL string, `url` as string, `description` as
  object**: all accepted (Postman schema allows them).
- **Postman v2.0 auth (key/value maps)** is normalised like v2.1 attribute lists.
- **File paths** in form-data/binary bodies come from the exporter's machine and are untrusted:
  imported file references are kept (for export) but always marked "re-pick needed" — they are
  never read or uploaded until Luis picks the file again on this machine — and are listed in the
  summary. This prevents a shared collection from uploading a local file (e.g., an SSH key) that
  happens to exist at the same path.
- **Untrusted archives and files**: zips are limited to 1 GB uncompressed in total (checked while
  extracting), entries with absolute paths, `..` segments or symbolic links are refused; single
  input files over 200 MB are refused; input must be UTF-8 (a BOM is ignored; CRLF accepted).
- **Secrets in collection files**: a Postman collection has no secret type; values stay as typed
  (plain), as in Postman.
- **Bruno secrets**: only names are available; they import as missing secrets.
- **Imported into a folder**: importing a collection into an existing folder makes the
  collection's top-level folders and requests children of that folder. The collection's auth,
  headers and scripts are applied to the target folder when it has none of its own; otherwise
  the imported items are placed in a new sub-folder named after the collection, which carries
  them. Collection variables are added to the containing collection (clashes per FR-018). The
  summary states which option was applied.
- **cURL with multiple URLs or unsupported protocols**: the first HTTP(S) URL is used and the
  rest listed in the notice; non-HTTP URLs are refused.

## Requirements *(mandatory)*

### Functional Requirements

**Import entry points**

- **FR-001**: Users MUST be able to open an Import dialog with Cmd/Ctrl+O and from the File menu
  and sidebar; it accepts a file picker selection, files/folders/zips dropped onto the window,
  and pasted text (JSON or cURL).
- **FR-002**: The dialog MUST detect the format automatically (Postman collection v2.0/v2.1,
  Postman environment, Postman globals, Postman data dump, Bruno `.bru` collection, Bruno YAML
  collection, cURL) and refuse anything else with a reason.
- **FR-003**: Before writing, the dialog MUST let the user choose the target workspace and
  location (new collection, or an existing collection/folder for collections and requests) and
  show what will be imported.

**Postman collections**

- **FR-004**: Import MUST map folders and requests preserving order and nesting; requests'
  method (any string), URL (`raw` preferred; rebuilt from parts when absent), query rows
  (disabled, description), path variables (values, descriptions), headers (disabled,
  description; raw-string form accepted), body (all modes incl. disabled body and raw language),
  auth (all types), scripts (prerequest → pre-request, test → post-response, at every level,
  disabled flag), collection variables (initial values, enabled, description), descriptions
  (string or object), examples, and `protocolProfileBehavior`.
- **FR-005**: Auth inheritance MUST be preserved: absent/null auth → Inherit; `noauth` → No auth.
  Types bearer, basic, apikey, oauth2, awsv4 and digest MUST map to Xenon's auth format (bearer
  and basic are sendable now; apikey, oauth2, awsv4 and digest become sendable in stage 5 and are
  shown as "available in a later stage" until then); other types MUST be stored unchanged, shown
  as "unsupported auth type (kept)", and sent without auth.
- **FR-006**: `protocolProfileBehavior.followRedirects`, `strictSSL` and `disableUrlEncoding` MUST
  map to request settings (follow redirects, SSL verification, encode URL); all
  `protocolProfileBehavior` keys MUST also be kept for export.
- **FR-007**: Every Postman field that has no Xenon equivalent (e.g., folder/request-level
  variables, proxy, certificate, script ids/names/packages, variable types, `_postman_id`,
  `_exporter_id`, `info.version`, unknown fields) MUST be preserved in the item's Postman
  passthrough so a later Postman export reproduces it.
- **FR-008**: Postman v2.0 collections MUST be accepted and normalised the same way as v2.1.

**Postman environments, globals, data dumps**

- **FR-009**: Environment JSON MUST import as a Xenon environment: plain values → initial
  values, `enabled` kept, `type: secret` → secret variables (non-empty values written to the
  vault, empty ones marked missing). No secret value may be written to a workspace file.
- **FR-010**: Globals JSON MUST merge into the target workspace's globals per FR-018.
- **FR-011**: A data dump (zip or folder) MUST list every collection and environment it contains
  with a per-item include checkbox and target workspace (default: current; "New workspace…"
  creates one inline), then import all selected items in one operation with one combined
  summary. Unrecognised parts MUST be skipped and listed.

**Bruno**

- **FR-012**: Bruno collections in `.bru` format and in YAML (OpenCollection) format MUST import:
  folders and requests in sequence order, method (incl. custom), URL, query/path params, headers
  (disabled rows), body modes (json, text, xml, form-urlencoded, multipart-form with files,
  file, graphql + vars), auth (inherit, none, basic, bearer, apikey, digest, oauth2, awsv4 map
  to Xenon; others kept and marked unsupported), docs → description, settings (encode URL,
  timeout, follow redirects), collection- and folder-level headers, auth and scripts.
- **FR-013**: Bruno environments MUST import as environments; secret variables MUST import as
  secrets marked missing.
- **FR-014**: Bruno scripts MUST be translated to `pm.*` using a published mapping of common
  `bru`/`req`/`res`/`test`/`expect` calls; untranslatable lines MUST be kept, prefixed with a
  `// XENON: untranslated` comment, and listed in the summary with file and line.
- **FR-015**: Bruno request variables and assertions MUST be converted to equivalent script lines
  (`pm.variables.set` in pre-request / post-response; `pm.test` + `pm.expect` per assertion
  operator); non-convertible ones MUST be listed.

**Common import behaviour**

- **FR-016**: Each import MUST produce a summary with counts per item kind and every warning
  (unsupported auth, packages/`pm.require`, missing files, variables with no Xenon place,
  untranslated script lines, skipped parts), each linking to the affected item. It MUST also list
  every `pm.*` API used by imported scripts that is outside the MVP script subset (PRD "What
  scripts can do"), grouped by API with the scripts that use it — this answers the PRD's open
  question about `pm` usage in Luis's collections.
- **FR-017**: Writing MUST be atomic per imported collection/environment: on failure nothing of
  that item remains; other items continue.
- **FR-018**: Name clashes (collection, environment, request in an existing folder, global or
  collection variable key) MUST offer Replace, Keep both (" (imported)" suffix; variables keep
  the existing value and record the imported value in the summary), or Skip, with "Apply to all".
- **FR-019**: Imports MUST show progress for more than 200 items, be cancellable before writing
  starts, and not block the UI.
- **FR-020**: Every imported item MUST get a new Xenon id; source ids are preserved for export.
- **FR-020a**: Imported form-data/binary file references MUST be marked "re-pick needed" and
  MUST NOT be read or sent until the user selects the file again; they are listed in the
  summary.
- **FR-020b**: Disabled scripts in the source (Postman `event.disabled`) MUST be imported as
  disabled scripts (stored, not run) and stay disabled on export.

**Saved examples**

- **FR-021**: A request MUST support any number of saved examples, each with: id, name, request
  snapshot (method, URL, params, path variables, headers, body, auth), response (status code,
  status text, headers, body, content type), and Postman passthrough fields (cookies, response
  time, preview language, original ids). Examples are stored with their request (same file).
- **FR-022**: Examples MUST be listed under their request in the sidebar (expandable, with status
  code), open in tabs following the request tab rules, be fully editable (request and response
  parts) and saved with Save, and support rename, duplicate and delete from the sidebar menu.
- **FR-023**: The response viewer MUST offer "Save response as example", creating an example from
  the request as sent (unresolved variables) and the received response; a body larger than 5 MB
  MUST still be stored and open in Raw first.

**cURL**

- **FR-024**: cURL import MUST support: `-X/--request`, `-H/--header`, `-d/--data`,
  `--data-raw`, `--data-ascii`, `--data-binary` (with `@file`), `--data-urlencode`, `-F/--form`
  (text and `@file`), `-G/--get`, `-I/--head`, `-T/--upload-file`, `--url`, `-u/--user`
  (+ `--basic`; `--digest` → Digest auth), `-b/--cookie` (→ Cookie header), `-A/--user-agent`,
  `-e/--referer`, `--json` (body + JSON Content-Type/Accept); `-k/--insecure` → SSL
  verification off; `-L/--location` → follow redirects on; `--compressed` and other no-effect
  flags ignored; unknown flags listed in a notice.
- **FR-025**: Bash (`\`), Windows cmd (`^`, `""`) and PowerShell (`` ` ``) line continuations
  and quoting MUST be normalised so all produce the same request.
- **FR-026**: Pasting text starting with `curl ` into the URL bar MUST replace the current tab's
  request with the parsed one (undoable); the import dialog MUST create a new saved request.

### Key Entities

- **Import source**: a detected input (kind, version, files) and its parsed items.
- **Import plan**: the items chosen, each with target workspace/location and clash resolution.
- **Import summary**: counts and warnings per item (kind, message, link to item).
- **Saved example**: id, name, request snapshot, response, passthrough; owned by a request.
- **Postman passthrough**: preserved source fields per collection/folder/request/example/variable
  for lossless Postman export.
- **Script translation report**: per Bruno script, translated lines and untranslated lines with
  locations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Importing the Postman fixture corpus (covering every field in FR-004–FR-008) yields
  100% of requests, folders, scripts, auth, variables and examples present, verified field by
  field; the same check is run on Luis's real exports before the stage is accepted.
- **SC-002**: A Postman collection of 1,000 requests imports in under 10 seconds with the UI
  responsive throughout.
- **SC-003**: Sorting a data dump of 50 collections and 20 environments into 4 workspaces takes
  one screen and one confirmation.
- **SC-004**: 0 secret values from Postman environments appear in any workspace file after import.
- **SC-005**: The same cURL command in bash, cmd and PowerShell syntax produces identical
  requests in 100% of the test corpus (≥ 30 commands).
- **SC-006**: At least 90% of lines in a corpus of real Bruno scripts translate without a flag;
  100% of untranslated lines are listed.
- **SC-007**: After import, every warning in the summary points to an existing item that can be
  opened in one click.

## Assumptions

- Postman's "Export collection" (v2.1 JSON) and data dump formats are the ones Luis uses; v3 YAML
  can be added later without changing this stage's model.
- Imported scripts are stored but not executed until stage 4; unsupported auth types become
  usable in stage 5; Postman export (stage 6) is where round-trip fidelity is proven end to end —
  this stage guarantees nothing needed for it is lost.
- Postman's `pm.require`/packages cannot work outside Postman; such scripts are imported as-is
  and flagged.
- Example request snapshots do not resolve variables (as in Postman).
- Bruno's translation mapping follows Bruno's own Bruno↔Postman translators as the reference.
