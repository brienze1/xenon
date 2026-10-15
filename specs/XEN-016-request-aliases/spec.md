# Feature Specification: Request Aliases

**Feature Branch**: `xen-016`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "Request aliases: named variations that live inside one saved request, so a request that is sent in a few fixed variations (log in as different test users; the same endpoint with a canary header; the same call on /v1, /v2, /v3) does not need separate requests or edit-and-revert. Each alias is a full, independent copy of the request's sendable content, stored with the request in its file. The sidebar is unchanged. A request tab shows an Aliases dropdown in its top-right corner (Default, every alias, Create new alias…). Source: https://github.com/brienze1/xenon/issues/19"

## Context

A request is often sent in a few fixed variations:

- **Login** as my user, a test user, a Mexican test user or a US test user.
- **Orders** with and without an `X-Canary: true` header.
- **Customers** on `/v1/customers`, `/v2/customers` and `/v3/customers`.

Today each variation is either a separate request (cluttering the sidebar) or the user edits the
request and changes it back every time. Aliases are named variations that live *inside* one
request: the request as it is today is its **Default**, and every alias is a complete,
independent version of it that can be picked from the request's tab and sent.

Postman has no equivalent. Constitution VI ("Postman parity, nothing more") puts anything outside
the PRD out of scope, so this feature is a **user decision** and the PRD records it as one
(Decisions, Oct 15, 2026).

### Decisions taken before this spec (user, 2026-10-15)

| Question (issue #19) | Decision |
| --- | --- |
| Storage model | **Full copy.** An alias stores its whole sendable content; a later change to the Default never reaches an alias. |
| What "Create new alias…" copies | **Always the saved Default**, whichever version is on screen. |
| Renaming an alias | Like XEN-012 click-to-rename: with an alias selected the breadcrumb shows one more crumb (the alias name); clicking it edits the name in place as a draft, saved with Cmd/Ctrl+S. |
| Deleting / duplicating an alias | A ⋯ menu on each alias row of the Aliases dropdown with **Delete only**. There is no Duplicate. |
| Tabs | **One tab per request** (an alias never opens a tab of its own); the dropdown switches the version shown in place; the tab remembers its alias across restarts; opening the request from the sidebar focuses the existing tab as it is; a newly opened tab starts on Default. |
| History | Each send records the alias it used; the History list shows `Login · mx-test-user`. |
| Postman export | Each alias is exported as **its own request** next to the Default, named `Login (mx-test-user)`. |
| AI access (MCP) | **Full:** AI clients can list, create, rename, delete and edit aliases, and send a given alias. |

### Settled from the issue and the code rather than asked

- **Unsaved edits when switching** — the issue's scope says "unsaved edits are tracked per
  alias": every version keeps its own draft, switching never prompts.
- **Names** — the dropdown tells versions apart only by name, and "Default" is already an entry
  in it, so alias names are unique within a request (ignoring case and surrounding spaces) and
  "Default" is reserved. Otherwise they follow the request name rule: 1–200 characters after
  trimming (`maxItemNameLength`, `internal/domain/entity/item_name.go`).
- **Examples** — the sidebar must stay unchanged and saved examples are listed under the request
  there, so examples belong to the request as a whole, not to an alias.
- **Name and description** — a request's name and description identify it in the sidebar, the
  tab and the documentation; an alias has its own name but no description.
- **Imports** — Postman, Bruno and cURL have no alias concept, so imports from them only create
  Defaults; a Xenon YAML export carries the request files as they are, so aliases survive a Xenon
  export → import.
- **Code snippets** — the snippet panel is generated from the version on screen, so it follows
  the selected alias.
- **Git sync** — aliases live in the request's file and are synced and committed with it; a request file changed on both sides is a file conflict, as today.

## Clarifications

### Session 2026-10-15

- Q: With an alias on screen, what does "Save As…" create? → A: A new request whose Default is the version on screen (with its unsaved edits) and which has no aliases; the original request and its aliases are unchanged.
- Q: How do aliases appear in a collection/folder Documentation (OpenAPI) tab? → A: Not at all — the document is generated from each request's Default, and applying a document never reads or changes aliases.
- Q: Does picking an alias in a preview tab pin it? → A: Yes — selecting, creating or deleting an alias pins a preview tab, as an edit does.
- Q: An alias with its own OAuth 2.0 auth — whose token does it use? → A: Its own. Each alias with its own OAuth 2.0 config has its own token slot (Get New Access Token, refresh, expiry are per alias); aliases that inherit auth use the folder/collection token as today; deleting an alias deletes its token.
- Q: Should scripts be able to save different variables per alias? → A: Yes, through per-alias scripts: each alias has its own pre-request and post-response scripts (e.g. the Default runs `pm.environment.set('token', …)`, alias `mx-test-user` runs `pm.environment.set('mxToken', …)`); the variables land in the normal scopes (local, environment, collection, globals). There is no alias-specific variable scope.
- Settled from the code (not asked): "Duplicate tab" (tab right-click menu) already opens a second tab on the same saved request (`duplicateTab`, `frontend/src/features/tabs/store.ts`), so "one tab per request" means aliases never open tabs of their own; a duplicated tab copies the version on screen and every version's drafts.
- Settled from the code (not asked): the MCP surface follows the existing resource tools — the `tree` tool's request reads include the aliases, the `tree_edit` tool's request target gains alias actions (create, rename, delete, edit), and the `send` tool takes an optional alias of a saved request (`internal/integration/entrypoint/controller/mcp_tree_edit_request.go`, `mcp_send_target.go`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create an alias and send it (Priority: P1)

Luis opens his **Login** request. In the top-right corner of the request tab an **Aliases**
dropdown shows **Default**. He opens it: it lists Default and, last, **Create new alias…**. He
picks Create new alias…, types `mx-test-user` and confirms. The tab now shows the alias: the
dropdown reads `mx-test-user`, the tab title reads `Login · mx-test-user`, and the breadcrumb
ends with an extra `mx-test-user` crumb. The request content is a copy of the saved Default. He
changes the body's `username` and `password` to the Mexican test account, presses Cmd+S, then
Send: the request goes out with the Mexican credentials. The sidebar still shows a single
**Login** row.

**Why this priority**: this is the whole point of the feature — one request, several named
variations, sent without editing and reverting.

**Independent Test**: create an alias on a saved request, edit its body, save, send; the send
carries the alias body, the request file holds the alias, and the sidebar shows one row.

**Acceptance Scenarios**:

1. **Given** a saved request with no aliases, **When** the user opens its tab, **Then** the
   Aliases dropdown shows "Default" and its menu lists Default and "Create new alias…".
2. **Given** a saved request, **When** the user picks "Create new alias…" and enters a valid
   name, **Then** the alias is written to the request's file as a copy of the saved Default,
   the tab switches to it, the tab title shows `<request> · <alias>` and the breadcrumb ends with
   the alias name.
3. **Given** a request whose Default has unsaved edits, **When** the user creates an alias,
   **Then** the alias is a copy of the **saved** Default (the unsaved edits stay a draft on the
   Default).
4. **Given** the tab shows an alias, **When** the user presses Send, **Then** the request sent
   is the alias's version (its method, URL, params, headers, body, auth, scripts and settings).
5. **Given** the Default's post-response script sets the environment variable `token` and alias
   `mx-test-user`'s sets `mxToken`, **When** each is sent, **Then** the Default's send sets only
   `token` and the alias's send sets only `mxToken`.
6. **Given** a request with aliases, **When** the sidebar is shown, **Then** it shows one row for
   the request and no alias rows; the request's saved examples are listed under it as before.
7. **Given** "Create new alias…" is picked, **When** the user enters an empty name, a name longer
   than 200 characters, "Default" (any case) or the name of an existing alias of the same
   request (ignoring case and surrounding spaces), **Then** nothing is created and the reason is
   shown inline in the name prompt.

---

### User Story 2 - Edit versions independently (Priority: P1)

The **Customers** request calls `/v1/customers`. Luis creates aliases `v2` and `v3` and changes
their URLs to `/v2/customers` and `/v3/customers`, saving each. Later he adds an `Accept-Language`
header to the Default and saves. `v2` and `v3` keep exactly what they had; the Default does not
pick up anything from them either.

**Why this priority**: independence is the contract that makes aliases trustworthy — without it
a save on one version silently changes another.

**Independent Test**: save different content in the Default and in an alias; reopen both from
disk; each holds only its own content.

**Acceptance Scenarios**:

1. **Given** a request with an alias, **When** the alias is edited and saved, **Then** the
   Default and every other alias are unchanged on disk.
2. **Given** a request with an alias, **When** the Default is edited and saved, **Then** every
   alias is unchanged on disk.
3. **Given** an alias, **When** the user changes any part of it — method, URL (including the path
   version), query params, path variables, headers, auto-header toggles, body, auth, scripts and
   "inherit parent scripts", settings — and saves, **Then** that change is stored for the alias
   and sent when the alias is sent.
4. **Given** the Default and an alias each have unsaved edits, **When** the user switches between
   them, **Then** no prompt appears and each version still shows its own unsaved edits.
5. **Given** any version of the request has unsaved edits, **When** the tab is shown, **Then** it
   shows the unsaved dot; **When** the user presses Cmd/Ctrl+S, **Then** only the version on
   screen is saved and the dot remains while another version still has unsaved edits.
6. **Given** any version has unsaved edits, **When** the user closes the tab, **Then** the usual
   Save / Don't Save / Cancel prompt appears; Save saves every version with unsaved edits, Don't
   Save discards them all.
7. **Given** the file changed on disk since the tab loaded it, **When** the user saves a version,
   **Then** the existing "changed on disk" banner appears, as for a request today.

---

### User Story 3 - Rename and delete aliases (Priority: P2)

Luis renamed his test account, so he selects alias `test-user`, clicks the `test-user` crumb in
the breadcrumb, types `qa-user` and presses Cmd+S: the dropdown, the crumb and the tab title now
read `qa-user`. He no longer needs `v3`, so he opens the dropdown, opens the ⋯ menu on the `v3`
row, picks Delete and confirms: `v3` is gone from the dropdown and from the file.

**Why this priority**: aliases accumulate; without rename and delete the dropdown rots. It is
secondary to creating and sending them.

**Independent Test**: rename an alias via the crumb and save; delete another via its ⋯ menu;
the file holds the new name and no deleted alias.

**Acceptance Scenarios**:

1. **Given** an alias is selected, **When** the user clicks its crumb (or focuses it and presses
   Enter), **Then** the name turns into an input with the name selected, following the XEN-012
   pattern: editing marks the tab dirty, Enter closes the input and keeps the draft, Escape
   cancels, Cmd/Ctrl+S saves the name with the alias's other pending edits.
2. **Given** an alias name draft that is empty, longer than 200 characters, "Default" or the name
   of another alias of the request (ignoring case and surrounding spaces), **When** the user
   saves, **Then** nothing is saved and the reason is shown inline at the name.
3. **Given** the Default is selected, **When** the breadcrumb is shown, **Then** it has no alias
   crumb and renaming the last crumb renames the request, as today.
4. **Given** the Aliases dropdown is open, **When** it lists aliases, **Then** each alias row has
   a ⋯ menu with **Delete** only, and the Default row has no ⋯ menu.
5. **Given** the user picks Delete on an alias, **When** they confirm, **Then** the alias is
   removed from the request's file and any unsaved edits on it are discarded; if it was the
   alias on screen, the tab switches to Default. **When** they cancel, nothing changes.

---

### User Story 4 - The tab remembers its alias (Priority: P2)

Luis leaves **Orders** on alias `canary` and quits Xenon. When he restarts, the Orders tab is
restored on `canary` (title `Orders · canary`). He clicks **Orders** in the sidebar: the tab is
focused as it is, still on `canary`. He closes the tab and opens Orders again from the sidebar:
the new tab starts on Default.

**Why this priority**: the selected alias is part of the tab's state, like its drafts; losing it
on restart sends the wrong version.

**Independent Test**: select an alias, restart, check the restored tab; close and reopen from
the sidebar, check it starts on Default.

**Acceptance Scenarios**:

1. **Given** a tab showing an alias, **When** Xenon restarts, **Then** the tab is restored on that
   alias with its unsaved edits per version, as drafts are restored today.
2. **Given** a request already open in a tab on an alias, **When** the user opens the request from
   the sidebar (single or double click), **Then** that tab is focused and stays on its alias.
3. **Given** a request not open in any tab, **When** the user opens it, **Then** the new tab shows
   the Default.
4. **Given** a restored tab whose alias no longer exists in the file (deleted elsewhere, e.g. by a
   git pull), **When** the tab is restored or the file reloads, **Then** the tab shows the
   Default.

---

### User Story 5 - History shows which alias was sent (Priority: P3)

After sending `Login` as `mx-test-user`, Luis opens the History view: the entry reads
`Login · mx-test-user`. A send of the Default reads just `Login`.

**Why this priority**: useful for telling sends apart, but sending works without it.

**Independent Test**: send the Default and an alias; History lists both, the alias entry with
its name.

**Acceptance Scenarios**:

1. **Given** a send of an alias, **When** it is recorded in History, **Then** the entry stores the
   alias's id and name, and the History list shows `<request> · <alias>`.
2. **Given** a send of the Default or of an unsaved tab, **When** it is recorded, **Then** the
   entry has no alias and is shown as today.
3. **Given** a History entry of an alias send, **When** the user reopens it, **Then** it opens as a
   draft of the request that was sent, as History entries open today.

---

### User Story 6 - Aliases travel with export, import and git (Priority: P3)

Luis exports his collection to Postman for a colleague: **Login** appears together with
**Login (test-user)** and **Login (mx-test-user)**, each a normal Postman request. He also
exports it as Xenon YAML and imports it in another workspace: the aliases are there. His git
sync pushes and pulls aliases with the request files.

**Why this priority**: data must not be lost at the boundaries, but day-to-day use does not
depend on it.

**Independent Test**: export a request with aliases to Postman and to Xenon YAML; inspect the
Postman file and import the Xenon file; commit/pull a change to an alias through git sync.

**Acceptance Scenarios**:

1. **Given** a request with aliases, **When** it is exported as Postman JSON, **Then** the export
   holds the Default as today and, right after it in the same folder, one request per alias named
   `<request> (<alias>)` with the alias's content; saved examples stay only on the Default's
   request; secrets follow the existing Postman export rules (never included).
2. **Given** a request with aliases, **When** it is exported as Xenon YAML and imported, **Then**
   the imported request has the same aliases with the same content.
3. **Given** a Postman, Bruno or cURL import, **When** it creates requests, **Then** they have no
   aliases.
4. **Given** git sync, **When** an alias is changed, created or deleted, **Then** the change is
   part of the request file's change; a request file changed on both sides is a file conflict
   resolved by choosing a side, exactly as for request files today.
5. **Given** a request without aliases, **When** it is saved, exported or synced, **Then** its
   file is byte-for-byte what it is today (no empty aliases key).

---

### User Story 7 - AI clients work with aliases (Priority: P3)

Luis asks his AI client to "send Login as mx-test-user". The client lists the aliases of
**Login** through Xenon's MCP server, sends the `mx-test-user` alias and reports the response.
Later it adds a `v4` alias to **Customers** with the URL changed to `/v4/customers`.

**Why this priority**: AI access is a Xenon feature (XEN-008) and the user chose full parity, but
it builds on the UI-facing behaviour above.

**Independent Test**: through the MCP tools, list aliases of a request, create/edit/rename/delete
one, and send one; check the file and History.

**Acceptance Scenarios**:

1. **Given** a request with aliases, **When** an AI client reads the request through the MCP
   tools, **Then** it sees the request's aliases (id and name) and can read each alias's content.
2. **Given** an AI client, **When** it creates, renames, deletes or edits an alias, **Then** the
   same rules as the UI apply (names, independence, a copy of the saved Default on create) and
   the change is recorded as AI activity like other AI edits.
3. **Given** an AI client, **When** it sends a request with an alias, **Then** the alias's
   version is sent, History records the alias and the AI source, and "Confirm before send" and
   approvals apply as for any AI send.
4. **Given** an AI client names an alias that does not exist on the request, **When** it sends or
   edits, **Then** the call is refused with a "not found" error naming the alias, and nothing is
   sent or written.

### Edge Cases

- **Request renamed** — alias names are unchanged; titles and History show the new request name
  with the alias name.
- **Request duplicated** — duplicating a request in the sidebar copies its aliases
  with it (the whole file is copied).
- **Request moved** — aliases move with the file.
- **Save As from an alias** — creates a new request from the version on screen, without aliases.
- **Documentation tab** — shows Defaults only; applying a document leaves aliases untouched.
- **Alias names that differ only by case or spaces** — refused as a clash.
- **Many aliases** — no limit; the dropdown lists them in creation order and scrolls.
- **Alias deleted while its tab shows it** (sidebar, git pull, MCP) — the tab switches to Default
  and its draft for that alias is dropped.
- **Unsaved tab (never saved)** — has no Aliases dropdown; aliases need a saved request.
- **Example tabs** — an example tab has no Aliases dropdown.
- **OAuth 2.0 on an alias** — an alias with its own OAuth 2.0 config has its own token; fetching
  one for the alias never replaces the Default's.
- **Duplicate tab on an alias** — the new tab shows the same alias with the same drafts.
- **Secrets** — an alias's auth and headers follow the same secret rules as the Default:
  `{{secret}}` variables are stored by name only, never by value.
- **File without aliases written by an older Xenon** — opens as a request with only Default.
- **File with aliases opened by an older Xenon** — out of scope (older builds are replaced by the
  auto-updater, XEN-013).

## Requirements *(mandatory)*

### Functional Requirements

**Storage and model**

- **FR-001**: A saved request MUST be able to hold any number of aliases, each with a stable id,
  a name and a full copy of the request's sendable content: method, URL, query params, path
  variables, headers, disabled auto-headers, body, auth, scripts, "inherit parent scripts" and
  settings.
- **FR-002**: Aliases MUST be stored in the request's own file, in creation order. A request with
  no aliases MUST be written exactly as today (no aliases key).
- **FR-003**: Saving the Default MUST NOT change any alias; saving an alias MUST NOT change the
  Default or any other alias. There is no fall-through from the Default to an alias.
- **FR-004**: A request's name, description, id, examples and file location belong to the request
  as a whole and are shared by the Default and all aliases.
- **FR-005**: Alias names MUST be trimmed, 1–200 characters, unique within the request ignoring
  case and surrounding spaces, and MUST NOT be "Default" in any case. A violating name MUST be
  refused with a message saying which rule failed, and nothing MUST be written.

**Managing aliases**

- **FR-006**: Users MUST be able to create an alias by name; the new alias MUST be a copy of the
  saved Default, written to the file immediately, and the tab MUST switch to it.
- **FR-007**: Users MUST be able to rename an alias through its breadcrumb crumb following the
  XEN-012 click-to-rename pattern (draft, dirty dot, saved with Cmd/Ctrl+S, inline errors on
  save).
- **FR-008**: Users MUST be able to delete an alias from the ⋯ menu on its dropdown row, after a
  confirmation; the alias is removed from the file immediately.
- **FR-009**: Creating, saving and deleting an alias MUST use the same "changed on disk" protection
  as saving a request: a write based on an older revision of the file is refused and the tab shows
  the existing banner. **Keep Your Version** writes the version on screen anyway; **Load from Disk**
  replaces only the version on screen with its content from disk. Either way the tab adopts the new
  revision for every version, and the other versions keep their unsaved drafts (each version's save
  only replaces its own part of the file).

**Request tab**

- **FR-010**: A saved request's tab MUST show an Aliases dropdown in its top-right corner listing
  Default, every alias by name in file order, and last "Create new alias…". Unsaved requests and
  example tabs MUST NOT show it. The dropdown, its ⋯ menus and both dialogs MUST be fully operable
  from the keyboard (arrows, Enter, Escape) and expose accessible names (the trigger reads
  "Aliases: <version on screen>"; each ⋯ button reads "Alias actions for <name>").
- **FR-011**: While an alias is selected, the tab title MUST be `<request name> · <alias name>`
  and the breadcrumb MUST end with the alias name as an extra crumb. While an alias is shown the
  request's own name (breadcrumb and side-panel Info) is plain text, so a tab never has two pending
  names; the request is renamed with Default on screen or from the sidebar.
- **FR-012**: Aliases MUST NOT open tabs of their own: opening a request from the sidebar opens or
  focuses its tab, and selecting an entry in the dropdown MUST switch the version shown in that tab.
  "Duplicate tab" MUST copy the version on screen and every version's unsaved drafts into the new tab.
- **FR-013**: Each version MUST keep its own unsaved edits; switching versions MUST NOT prompt or
  discard. The tab is dirty while any version has unsaved edits. Cmd/Ctrl+S MUST save the version
  on screen; the close prompt's Save MUST save every dirty version (Default first, then aliases in
  file order), and Don't Save MUST discard all of them. If one of those saves fails, the saving
  stops, the versions already saved stay saved, and the tab stays open showing that error.
- **FR-014**: The tab MUST remember its selected alias and per-version drafts across restarts.
  Opening an already-open request from the sidebar MUST focus its tab unchanged; a newly opened
  tab MUST start on Default. A selected alias that no longer exists MUST fall back to Default.
- **FR-014a**: Selecting, creating or deleting an alias in a preview tab MUST pin the tab.
- **FR-014b**: "Save As…" with an alias on screen MUST create a new request whose Default is the version on screen
  (including its unsaved edits) and which has no aliases; the original request and its aliases MUST stay unchanged.

**Sending and related views**

- **FR-015**: Send MUST send the version on screen, including its unsaved edits, as a send works
  today.
- **FR-015a**: An alias with its own OAuth 2.0 auth MUST have its own token (fetched, refreshed,
  expired and cleared independently of the Default and other aliases); an alias that inherits auth
  MUST use the inherited token as today. Deleting an alias MUST delete its token.
- **FR-015b**: Sending an alias MUST run the alias's own pre-request and post-response scripts
  (with the alias's "inherit parent scripts" setting), never the Default's, so each version's
  scripts can set different variables in the existing scopes. No alias-specific variable scope is
  added.
- **FR-016**: The code snippet panel MUST generate from the version on screen.
- **FR-017**: A History entry of a send MUST store the id and name of the alias sent (none for the
  Default), and the History list MUST show `<request name> · <alias name>` for alias sends.

**Boundaries**

- **FR-018**: Postman JSON export MUST write each alias as a separate request placed right after
  its Default in the same folder, named `<request name> (<alias name>)`, with the alias's content
  and no examples. The name is used as is even if a sibling already has it (Postman allows duplicate
  names). Importing that export back creates separate requests, not aliases.
- **FR-019**: Xenon YAML export and import MUST carry aliases unchanged.
- **FR-020**: Postman, Bruno and cURL imports MUST create requests without aliases.
- **FR-021**: Git sync MUST commit, push and pull aliases as part of the request file, with the
  request-file merge behaviour of today (a file changed on both sides is a file conflict resolved
  by choosing a side).
- **FR-022**: Duplicating or moving a request MUST carry its aliases.
- **FR-022a**: A collection's or folder's Documentation (OpenAPI) document MUST be generated from each request's
  Default only, and applying a document MUST NOT read, change or remove aliases.

**AI access (MCP)**

- **FR-023**: The MCP tools MUST expose a request's aliases (id, name, content) when the request
  is read, through the same leak guard as the Default (secret values never appear in an AI result).
- **FR-024**: The MCP tools MUST let AI clients create, rename, delete and edit an alias under
  FR-003, FR-005 and FR-006, recorded as AI activity.
- **FR-025**: The MCP send tool MUST accept an optional alias of a saved request and send that
  alias's version, recording the alias in History; an unknown alias MUST be refused with a
  not-found error.

**Product record**

- **FR-026**: The PRD MUST record request aliases as a user decision (Decisions, Oct 15, 2026),
  since Postman has no equivalent.

### Key Entities

- **Request**: a saved request as today (id, name, description, examples, location) whose own
  sendable content is the **Default**, plus an ordered list of aliases.
- **Alias**: a named variation of one request — id (unique within the request), name, and a full
  copy of the sendable content (method, URL, params, path variables, headers, disabled
  auto-headers, body, auth, scripts, inherit-scripts, settings).
- **Request tab state**: the selected version (Default or an alias id) and an unsaved draft per
  version.
- **History entry**: as today, plus the alias id and name of the version sent (empty for the
  Default).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can turn one request into three variations (e.g. v1/v2/v3) and send each one
  without the sidebar gaining any row.
- **SC-002**: Switching between two saved versions of a request and sending takes no more than
  two clicks plus Send, and never requires editing the request.
- **SC-003**: In 100% of save operations on one version, every other version of the request is
  unchanged on disk.
- **SC-004**: A request without aliases produces a file identical to today's for 100% of saves,
  exports and syncs.
- **SC-005**: After a restart, 100% of request tabs reopen on the version they showed before.
- **SC-006**: A Postman export of a request with N aliases yields N + 1 requests, none of which
  contain a secret value.

## Assumptions

- The selected alias and per-version drafts are kept in the same per-workspace tab state that
  restores tabs today.
- Alias ids are generated like other item ids and never shown to the user.
- No limit on the number of aliases per request is needed; real use is a handful.
- The request's collection/folder inheritance (auth, headers, scripts) applies to every version
  the same way it applies to the Default today.
