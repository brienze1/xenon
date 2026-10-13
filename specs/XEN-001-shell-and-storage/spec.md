# Feature Specification: Shell and Storage (Stage 1)

**Feature Branch**: `xen-001`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 1 — Shell and storage: workspaces, collections, folders, requests stored as plain files, sidebar, tabs, request builder (params, headers, body, Bearer and Basic auth), sending, response viewer, Postman-style window layout in Xenon's visual identity. Source: docs/prd/xenon-mvp-prd.md (Build order, item 1)."

## Context

Xenon is a local-first desktop API client that replaces Postman for one user (Luis) who works
across 4 companies. This is the first of seven build stages; at the end of it, simple HTTP
calls work end to end and are saved as plain files. Later stages add environments and secrets
(2), Postman/Bruno import (3), scripts and console (4), daily comfort (5), export (6) and git
sync (7). The stored data written in this stage is the format every later stage extends, so it
must already have room for what those stages add.

## Clarifications

### Session 2026-10-09

- Q: Postman's sidebar History tab is not in the PRD — include it? → A: Yes, add History in
  stage 1 (local per workspace, never synced, re-openable in a tab).
- Q: How are form-data/binary file paths stored in request files? → A: Relative to the
  workspace folder when the file is inside it; absolute otherwise, flagged "missing" when not
  found.
- Q: A request file changes on disk while its tab has unsaved edits — what happens? → A: Tabs
  without edits reload silently; a tab with edits shows a banner offering "Keep mine" or "Load
  from disk".

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send a request and read the response (Priority: P1)

Luis opens Xenon, opens a new tab, picks a method, types a URL, adds query params, headers, a
body and Bearer or Basic auth, presses Send, and reads the response: status, time, size, body
(formatted), and headers. While the request is in flight he can cancel it.

**Why this priority**: This is the core loop of an API client. Without it nothing else matters,
and on its own it already replaces Postman for ad-hoc calls.

**Independent Test**: Start Xenon with no workspace content, open a new tab, send a request to a
known endpoint, and verify every part of the response view. Delivers value with no saving at
all.

**Acceptance Scenarios**:

1. **Given** a new empty tab, **When** Luis sets method GET and URL
   `https://api.example.test/users?page=2` and presses Send, **Then** the response view shows
   the status code and text coloured by class, total time, response size, and the body.
2. **Given** a URL with query string `?page=2&limit=10`, **When** Luis looks at the Params tab,
   **Then** it lists `page=2` and `limit=10`; **and When** he unticks `limit` or edits a value
   in the table, **Then** the URL updates to match; **and When** he edits the URL, **Then** the
   table updates to match.
3. **Given** a URL `https://api.example.test/users/:id`, **When** Luis opens Params, **Then** a
   separate Path Variables table shows `id`, and the value he enters replaces `:id` when sent.
4. **Given** a request with method POST and a raw JSON body, **When** Luis sends it, **Then**
   the request carries the body and a `Content-Type: application/json` header that is shown
   greyed in the Headers tab as auto-generated, and that he can untick to suppress.
5. **Given** a request with Bearer auth token `abc`, **When** sent, **Then** the request carries
   `Authorization: Bearer abc`; with Basic auth `user`/`pass`, **Then** it carries the Basic
   header for those credentials.
6. **Given** a request is in flight, **When** Luis presses Cancel (the Send button while
   sending), **Then** the request is aborted and the response area says it was cancelled.
7. **Given** a response arrived, **When** Luis hovers the time, **Then** he sees a breakdown of
   DNS lookup, connection, TLS handshake, time to first byte and download.
8. **Given** a JSON, XML or HTML response, **When** Luis views Body → Pretty, **Then** it is
   formatted and highlighted with collapsible sections; Raw shows it exactly as received;
   Preview renders HTML or shows an image.
9. **Given** a response body, **When** Luis searches in it, copies it, or saves it to a file,
   **Then** matches are highlighted, the exact body is on the clipboard, or the exact bytes are
   written to the chosen file.
10. **Given** a response larger than 5 MB, **When** it arrives, **Then** the Body opens in Raw
    view first and Luis can still switch to Pretty.
11. **Given** the target is unreachable, the request times out, or TLS verification fails,
    **When** Luis sends, **Then** the response area shows a clear error naming what failed
    (for example "Could not resolve host", "Timed out after 5000 ms", "Certificate not
    trusted — turn off SSL verification in Settings to send anyway").

---

### User Story 2 - Organise requests in a workspace that lives as files (Priority: P1)

Luis creates a workspace for a company, choosing a folder anywhere on disk (for example inside
that company's git repo). He creates collections, nested folders and requests, renames,
duplicates and deletes them from the sidebar or its right-click menu, and searches by name or
URL. Everything is stored as readable files in that folder, one file per request, in folders
that mirror the sidebar, so he can commit it by hand.

**Why this priority**: Saving and organising is what makes Xenon a daily tool rather than a
scratchpad, and the on-disk files are the product's core promise.

**Independent Test**: Create a workspace in an empty folder, build a small tree, then inspect
the folder on disk and restart the app: the files match the tree and the tree comes back
identical.

**Acceptance Scenarios**:

1. **Given** no workspaces, **When** Luis creates workspace "Acme" in folder `~/work/acme`,
   **Then** the folder contains a workspace file and empty collections/environments folders,
   and the sidebar shows an empty "Acme" with an empty-state illustration.
2. **Given** a folder that already contains a Xenon workspace (for example cloned from git),
   **When** Luis chooses "Open workspace folder", **Then** it is added to his workspace list
   with all its collections, folders and requests.
3. **Given** a workspace, **When** Luis creates collection "Billing API", folder "Invoices"
   inside it, folder "Drafts" inside that, and request "Create invoice" in "Drafts", **Then**
   the sidebar shows that tree and the disk has one folder per collection/folder and one file
   per request in the matching place.
4. **Given** items in the sidebar, **When** Luis renames, duplicates or deletes one (from the
   item's right-click menu or the sidebar's own actions), **Then** the sidebar and the files on
   disk change accordingly; deleting asks for confirmation and removes children too; a
   duplicate is named "<name> Copy" and placed right after the original.
5. **Given** the order of items in a folder, **When** the app restarts, **Then** the order is
   the same (order is stored in the files, not derived from names).
6. **Given** many requests, **When** Luis types in the sidebar search, **Then** only requests
   whose name or URL contains the text (case-insensitive) are shown, with their parent
   folders, within the current workspace only.
7. **Given** a collection or folder, **When** Luis opens its settings, **Then** he can set its
   Authorization (inherit, none, Bearer, Basic) and headers; requests whose auth is "Inherit
   from parent" use the nearest ancestor's auth, and ancestor headers are sent unless the
   request defines a header with the same name.
8. **Given** two requests with the same name in the same folder, **When** both are saved,
   **Then** both exist (names need not be unique) and each has its own file.
9. **Given** a request name containing characters that are not allowed in file names on some
   operating system (for example `GET /users/:id?`), **When** it is saved, **Then** the file
   is created with a safe name on macOS, Windows and Linux, and the sidebar still shows the
   original name.

---

### User Story 3 - Work in tabs like Postman (Priority: P2)

Luis opens requests in tabs with Postman's behaviour: a single click opens a preview tab
(italic title) that the next single click replaces; editing it or double-clicking pins it. He
opens new empty tabs with + or Cmd/Ctrl+T, sends them without saving and saves them into a
collection later. Unsaved changes show a dot; closing a changed tab asks save, discard or
cancel. Tabs come back per workspace after a restart.

**Why this priority**: Tabs are how Luis juggles several calls at once; Postman muscle memory
depends on these exact behaviours.

**Independent Test**: With a saved tree, exercise preview/pin, dirty state, close prompts,
save-into-collection and restart restore, all without touching workspaces or sending.

**Acceptance Scenarios**:

1. **Given** no preview tab, **When** Luis single-clicks request A, **Then** A opens in a
   preview tab with an italic title; **When** he single-clicks request B, **Then** B replaces A
   in that same tab.
2. **Given** A is in a preview tab, **When** Luis edits it or double-clicks it in the sidebar or
   on the tab, **Then** the tab becomes pinned (normal title) and the next single click opens a
   new preview tab.
3. **Given** a request already open in a tab, **When** Luis clicks it in the sidebar, **Then**
   that tab is focused instead of opening a second one.
4. **Given** a pinned tab with edits, **Then** it shows an unsaved dot; **When** Luis presses
   Cmd/Ctrl+S, **Then** the file is written and the dot disappears.
5. **Given** a tab with unsaved changes, **When** Luis closes it, **Then** he is asked to Save,
   Don't save, or Cancel, and the chosen action happens.
6. **Given** a new unsaved tab, **When** Luis presses Save (or Save as, Cmd/Ctrl+Shift+S, on a
   saved request), **Then** he picks a name and a collection or folder in the current workspace,
   and the request is saved there and the tab now points to it.
7. **Given** open tabs (including unsaved new tabs and unsaved edits) in workspace Acme, **When**
   the app is closed and reopened, **Then** Acme's tabs, their order, the active tab and their
   unsaved edits are restored.
8. **Given** a tab's request was deleted from the sidebar, **Then** its tab closes; if the tab
   had unsaved edits, Luis is first asked whether to keep them as a new unsaved tab.
9. **Given** several tabs, **When** Luis drags a tab, **Then** the tab order changes and is
   remembered.

---

### User Story 4 - Keep companies apart (Priority: P2)

Luis switches workspaces from the switcher at the top left. Switching swaps the whole sidebar,
the open tabs and every remembered state. Nothing from one company appears in another. He can
create, rename and remove workspaces, with no limit on their number.

**Why this priority**: Clean separation of 4 companies is one of the main reasons Xenon exists,
but it only matters once there is content to separate.

**Independent Test**: Create two workspaces with different content and tabs, switch back and
forth, restart, and verify each keeps exactly its own content, tabs and layout.

**Acceptance Scenarios**:

1. **Given** workspaces Acme and Globex with their own collections and tabs, **When** Luis
   switches from Acme to Globex, **Then** the sidebar, tabs, active tab, sidebar expansion and
   search show only Globex's; **When** he switches back, **Then** Acme's state is exactly as he
   left it, including unsaved edits.
2. **Given** a workspace, **When** Luis renames it, **Then** its display name changes everywhere
   and its folder on disk is not moved.
3. **Given** a workspace, **When** Luis removes it, **Then** he chooses between "Remove from
   Xenon" (files stay on disk) and "Remove and move the folder to the trash", confirms, and the
   workspace disappears from the switcher; Xenon then switches to another workspace or shows
   the welcome screen if none are left.
4. **Given** the app restarts, **Then** it reopens the workspace that was active when it closed.
5. **Given** a workspace whose folder was moved or deleted outside Xenon, **When** Luis selects
   it, **Then** Xenon says the folder is missing and offers "Locate folder…" or "Remove from
   Xenon".

---

### User Story 5 - Find a request I sent earlier (Priority: P3)

The sidebar has a History view next to Collections, as in Postman. Every request Luis sends in a
workspace is listed there, newest first and grouped by day, with method, URL, status and time.
Clicking an entry opens that request, exactly as it was sent, in a new tab.

**Why this priority**: Luis uses Postman's history daily to get back to ad-hoc calls he never
saved; useful but not needed for the first sends.

**Independent Test**: Send several requests (saved and unsaved) in two workspaces, restart, and
verify each workspace's History lists only its own sends and re-opens them faithfully.

**Acceptance Scenarios**:

1. **Given** Luis sends three requests in Acme, **When** he opens the History view, **Then** it
   lists the three newest first under "Today", each with method, URL, response status (or the
   error) and time.
2. **Given** a History entry, **When** Luis clicks it, **Then** a new unsaved tab opens with the
   request as it was sent (method, URL, params, headers, body, auth, settings), which he can send
   again or save into a collection.
3. **Given** History in Acme, **When** Luis switches to Globex, **Then** Globex's History shows
   only Globex's sends.
4. **Given** History entries, **When** Luis searches the sidebar while History is shown, **Then**
   entries filter by URL or method; **and When** he chooses "Clear history" (with
   confirmation), **Then** the workspace's history is emptied.
5. **Given** the app restarts, **Then** History is still there, and it is never written into the
   workspace folder.

---

### User Story 6 - A familiar, comfortable window (Priority: P3)

The window follows Postman's placement: workspace switcher top left, sidebar on the left, tabs
with the request above the response in the main area, a collapsible panel on the right, a
console drawer at the bottom (hidden by default) and a footer. The sidebar, right panel and
console can each be hidden; their sizes are remembered. The request/response split can be
dragged and switched to side by side. Colours and fonts are Xenon's own dark neon identity.

**Why this priority**: Comfort for a full working day, but every earlier story works without
the polish.

**Independent Test**: Resize, hide and toggle every region, restart, and verify the layout is
restored; compare colours against the PRD palette.

**Acceptance Scenarios**:

1. **Given** the default window, **Then** regions are placed as described, the console drawer is
   hidden, and the right panel shows the Info view (request name, markdown description and its
   location in the collection) for the active request.
2. **Given** Luis resizes or hides the sidebar, right panel or console, or drags the
   request/response split, or switches it to side by side, **When** the app restarts, **Then**
   the layout is the same, separately for each workspace.
3. **Given** any screen, **Then** colours come only from the PRD palette, neon is used for
   accents and state (never for long text), glow appears only on active or attention elements
   (Send button, active tab underline, focused input, running request), HTTP methods use the
   PRD method colours, and empty states (no workspace, no requests, no response yet) show the
   graffiti-style illustration with the marker display font.
4. **Given** Luis presses Cmd/Ctrl+Enter, Cmd/Ctrl+S, Cmd/Ctrl+Shift+S, Cmd/Ctrl+T or
   Cmd/Ctrl+W, **Then** Send, Save, Save as, New tab or Close tab happens (Cmd on macOS, Ctrl
   on Windows and Linux).

---

### Edge Cases

- **Malformed or hand-edited file**: a request/collection/folder file that cannot be read is
  shown in the sidebar with an error marker and the reason; the rest of the workspace loads
  normally, and Xenon never overwrites the broken file unless Luis saves over it explicitly.
- **Files changed outside Xenon** (hand edits, `git pull`, another editor): see FR-017.
- **Body file paths**: a form-data/binary file inside the workspace folder is stored relative to
  the folder (works after a clone on another machine/OS); a file outside is stored as an absolute
  path. A referenced file that does not exist is shown as "missing" in the body editor and
  blocks sending until it is re-picked or removed.
- **History growth**: history keeps the most recent 1,000 sends per workspace; older entries are
  dropped automatically.
- **Unknown fields** in a file (for example written by a newer Xenon version or a later stage)
  are preserved untouched when Xenon saves that file.
- **Crash or power loss during save**: files are never left half-written; either the old or the
  new content is on disk.
- **Folder not writable** (permissions, read-only disk): the save fails with a clear message and
  the tab keeps its unsaved edits.
- **Two workspaces pointing at the same folder**, or a workspace folder inside another
  workspace's folder: refused with an explanation when creating or opening.
- **Very large response** (hundreds of MB): Raw view stays responsive; Pretty is offered but
  may warn that formatting a body this size can be slow.
- **Binary response** (image, PDF, zip): Pretty is not offered; Raw shows that it is binary and
  its size; Preview renders images; Save to file writes the exact bytes.
- **Form-data or binary body file missing at send time**: send is blocked with a message naming
  the missing file.
- **Custom method** typed in (for example `PURGE`) is sent as typed, upper-cased.
- **Redirects** when "follow redirects" is off: the 3xx response itself is shown.
- **Empty states**: no workspace (welcome screen with create/open actions), empty workspace,
  empty collection, no response yet.
- **Trash unavailable or failing** (e.g., a folder on a volume without a trash): removal is
  aborted, the workspace stays registered, and the message offers "Remove from Xenon" (keep
  files) instead.
- **Symbolic links** inside a workspace folder are ignored, so nothing outside the folder is
  read or written.
- **Name collisions on disk** when two siblings have the same name or names that only differ by
  case (relevant on case-insensitive file systems): each still gets its own file.

## Requirements *(mandatory)*

### Functional Requirements

**Workspaces**

- **FR-001**: Users MUST be able to create a workspace by giving it a name and choosing a folder
  anywhere on disk; Xenon writes the workspace's files into that folder.
- **FR-002**: Users MUST be able to open an existing workspace folder and add it to their
  workspace list.
- **FR-003**: Users MUST be able to rename a workspace (display name only) and remove it, either
  keeping its folder or moving the folder to the OS trash, after confirmation.
- **FR-004**: Users MUST be able to switch workspaces from a switcher at the top left. Switching
  MUST replace the sidebar, open tabs, active tab, sidebar expansion/search state and layout
  with the target workspace's own, and nothing from one workspace may be visible or reachable
  from another.
- **FR-005**: Each workspace MUST remember its open tabs (order, active tab, preview/pinned,
  unsaved edits and unsaved new tabs), sidebar state (expanded nodes, selection) and layout
  (region sizes, visibility, split orientation) across restarts. This state MUST be stored
  outside the workspace folder, so it never ends up in the workspace's git repository.
- **FR-006**: There MUST be no limit on the number of workspaces. The app MUST reopen the last
  active workspace on start.

**Collections, folders, requests and storage**

- **FR-007**: A workspace MUST hold collections; a collection MUST hold folders and requests;
  folders MUST nest to any depth.
- **FR-008**: Users MUST be able to create, rename, duplicate and delete collections, folders and
  requests from the sidebar and from each item's right-click menu. Deleting MUST ask for
  confirmation and delete children.
- **FR-009**: Each workspace MUST be stored as plain, human-readable YAML in its folder with this
  shape: a workspace file at the root, an `environments/` folder, and a `collections/` folder in
  which every collection and folder is a directory (with its own settings file) and every
  request is its own file, so the directory tree mirrors the sidebar.
- **FR-010**: The order of collections, folders and requests MUST be stored in the files and
  restored exactly.
- **FR-011**: File and directory names MUST be derived from item names in a way that is valid on
  macOS, Windows and Linux and stays unique among siblings (including on case-insensitive file
  systems), while the original display name is stored inside the file.
- **FR-012**: Every stored file MUST carry a format version, and the stored request, folder,
  collection and workspace formats MUST already have places for data later stages add
  (variables, pre-request and post-response scripts, "inherit parent scripts", the other
  authorization types, request description, Postman-specific fields kept for round trip) so
  that later stages extend the format without rewriting existing files.
- **FR-013**: Saving MUST be atomic per file (never half-written) and MUST preserve fields Xenon
  does not understand.
- **FR-014**: Collections and folders MUST have their own Authorization (inherit, none, Bearer,
  Basic) and headers. A request with "Inherit from parent" auth MUST use the nearest ancestor
  that defines auth (a collection's "inherit" means no auth). Headers defined on ancestors MUST
  be sent with the request unless the request (or a nearer ancestor) defines a header of the
  same name (case-insensitive), in which case the nearer one wins.
- **FR-015**: Sidebar search MUST filter requests by name or URL (case-insensitive substring)
  within the current workspace, keeping their parent folders visible.
- **FR-016**: A file that cannot be parsed MUST appear in the sidebar as an error item with the
  reason, without blocking the rest of the workspace, and MUST NOT be overwritten unless the user
  explicitly saves over it.
- **FR-017**: Changes made to workspace files outside Xenon while it is open MUST be picked up
  without a restart: the sidebar refreshes, and open tabs without unsaved edits reload silently.
  If a tab has unsaved edits for a request that changed on disk, the tab MUST show a banner
  offering "Keep mine" (the next save overwrites the disk version) or "Load from disk" (discards
  the edits). If the file was deleted on disk, the banner offers "Keep as unsaved tab" or "Close".

**Tabs**

- **FR-018**: A single click on a request in the sidebar MUST open it in a preview tab (italic
  title) that the next single click replaces; editing the request or double-clicking it (in the
  sidebar or on the tab) MUST pin the tab. Clicking a request that is already open MUST focus
  its tab.
- **FR-019**: Users MUST be able to open a new empty tab with a + button and Cmd/Ctrl+T, send it
  without saving, and save it into a chosen collection or folder of the current workspace.
- **FR-020**: A tab with unsaved changes MUST show a dot. Closing it MUST ask Save / Don't save /
  Cancel. Users MUST be able to Save (Cmd/Ctrl+S) and Save as (Cmd/Ctrl+Shift+S, which creates a
  copy at a chosen place and switches the tab to the copy).
- **FR-021**: Tabs MUST be reorderable by dragging.
- **FR-022**: When a request with an open tab is deleted, its tab MUST close; if it had unsaved
  edits, the user MUST first be offered to keep them as an unsaved new tab.

**Request builder**

- **FR-023**: The request bar MUST offer methods GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS and
  a custom method typed by the user, a URL field, a Send button that turns into Cancel while the
  request is in flight, and Save.
- **FR-024**: The request MUST have sub-tabs in Postman's order: Params, Authorization, Headers,
  Body, Scripts, Settings. In this stage the Scripts tab is present but shows that scripts arrive
  in a later stage.
- **FR-025**: Params MUST show the URL's query parameters as a key/value table kept in sync with
  the URL in both directions, and path variables (`:name` segments) in a separate table whose
  values are substituted into the URL when sending.
- **FR-026**: Authorization MUST support Inherit from parent, No auth, Bearer token and Basic
  auth in this stage.
- **FR-027**: Headers MUST be a key/value table with per-row enable checkbox and description. The
  headers Xenon adds automatically (Content-Type from the body type, Content-Length, Host,
  User-Agent, Accept, Accept-Encoding, Connection, and Authorization from auth) MUST be listed
  greyed out, and each MUST be possible to switch off.
- **FR-028**: Body MUST support: none; form-data (text fields and files); x-www-form-urlencoded;
  raw with type JSON, XML, Text, HTML or JavaScript and matching syntax highlighting; binary
  (a single file); GraphQL (query plus JSON variables, sent as a JSON POST body).
- **FR-028a**: A file referenced by a form-data row or binary body MUST be stored relative to the
  workspace folder when it is inside it and as an absolute path otherwise; a file that cannot be
  found MUST be marked "missing" and block sending until fixed.
- **FR-029**: Settings MUST offer per request: timeout in milliseconds (0 = no timeout, the
  default), follow redirects (default on), SSL certificate verification (default on) and encode
  URL automatically (default on).
- **FR-030**: Every key/value table (params, path variables, headers, form-data, urlencoded,
  collection/folder headers) MUST support per-row enable/disable and a bulk-edit mode where the
  rows are edited as text (`key:value` per line, `//` prefix for disabled rows), and MUST accept
  `{{variable}}` text in keys and values (resolution arrives in stage 2; until then it is sent
  literally).

**Sending and response**

- **FR-031**: Sending MUST perform the HTTP request in the backend with the effective settings,
  inherited auth and headers, and MUST measure DNS lookup, connect, TLS handshake, time to first
  byte, download and total time.
- **FR-032**: An in-flight request MUST be cancellable; cancelling MUST abort the network call.
- **FR-033**: The response view MUST show a status line (code and text coloured by class: 2xx
  green, 3xx yellow, 4xx/5xx red), total time with a hover breakdown of the timings in FR-031,
  and response size.
- **FR-034**: The response Body tab MUST offer Pretty (formatted, highlighted, foldable JSON, XML
  and HTML), Raw (exact text) and Preview (rendered HTML or image) views, search within the body,
  copy body and save the body to a file (exact bytes).
- **FR-035**: Responses larger than 5 MB MUST open in Raw first.
- **FR-036**: The response Headers tab MUST list response headers as a table, including repeated
  headers.
- **FR-037**: Errors that prevent a response (DNS failure, connection refused, timeout, TLS
  verification failure, cancelled, missing body file, invalid URL) MUST be shown in the response
  area with a message that names the cause and, where there is one, the setting that changes it.
- **FR-038**: The last response of each open tab MUST stay available while the app is running
  (switching tabs or workspaces does not lose it). Responses are not saved to disk in this stage.

**History**

- **FR-038a**: Every send (saved or unsaved request, success or error) MUST be recorded in the
  current workspace's History with: time sent, the request as sent (method, URL, params, headers,
  body, auth, settings — as written by the user, before any variable resolution), and the result
  summary (status code and text, or the error; total time; size). Response bodies are not kept.
- **FR-038b**: The sidebar MUST offer a History view (next to Collections) listing entries newest
  first, grouped by day; clicking an entry opens a new unsaved tab with that request.
- **FR-038c**: History MUST be per workspace, stored in local app data (never in the workspace
  folder, never synced), searchable by URL or method, clearable with confirmation, and capped at
  the 1,000 most recent entries per workspace.

**Window, identity and shortcuts**

- **FR-039**: The window MUST follow Postman's placement: workspace switcher top left, sidebar
  left, tab strip and request above response in the main area, a collapsible right side panel
  (showing Info — name, markdown description, location — in this stage), a bottom console drawer
  (hidden by default; its content arrives in stage 4) and a footer.
- **FR-040**: The sidebar, right panel and console MUST each be hideable and resizable; the
  request/response split MUST be draggable and switchable between stacked and side by side. All
  of these MUST be remembered per workspace.
- **FR-041**: The UI MUST use only the PRD palette, method colours, fonts (Inter for UI,
  JetBrains Mono for code/URLs/bodies, a marker display font for the wordmark, welcome screen and
  empty states) and neon/glow rules; dark theme only.
- **FR-042**: Shortcuts in this stage: Send (Cmd/Ctrl+Enter), Save (Cmd/Ctrl+S), Save as
  (Cmd/Ctrl+Shift+S), New tab (Cmd/Ctrl+T), Close tab (Cmd/Ctrl+W). Cmd on macOS, Ctrl on Windows
  and Linux.
- **FR-043**: The app MUST run on macOS, Windows and Linux with the same behaviour, using the
  Xenon app icon.

### Key Entities

- **Workspace**: a company's space. Display name, folder location on disk, format version, and
  (later) globals and sync settings. Contains collections and environments. Hard boundary.
- **Workspace list (local)**: the user's list of known workspaces and which one is active; kept
  in Xenon's local app data, not in any workspace folder.
- **Workspace UI state (local)**: per workspace — open tabs (with unsaved drafts), active tab,
  sidebar expansion/selection and active sidebar view (Collections/History), layout sizes and
  visibility. Local app data only.
- **Collection**: top-level group in a workspace. Name, description, order, auth, headers, and
  (later) variables and scripts. Contains folders and requests.
- **Folder**: group inside a collection or another folder. Same settings as a collection plus
  (later) "inherit parent scripts". Contains folders and requests.
- **Request**: name, description, method, URL, query params, path variables, headers, body (by
  type), auth, settings, order, and (later) scripts and "inherit parent scripts".
- **Key/value row**: key, value, enabled, description (and for form-data, text or file).
- **Response (in memory)**: status code and text, headers, body bytes, size, timings, or an
  error.
- **History entry (local)**: per workspace — sent-at time, the request as written by the user,
  and the result summary (status or error, total time, size). Local app data only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a cold app start, Luis can open a new tab and send a GET request in under 10
  seconds and 4 interactions (new tab, type URL, Send).
- **SC-002**: A workspace with 1,000 requests across 4 collections opens and shows its sidebar in
  under 2 seconds, and sidebar search updates while typing (under 200 ms per keystroke).
- **SC-003**: A 50 MB response is displayed (Raw) within 2 seconds of being received, and the UI
  stays responsive (scrolling and typing elsewhere are not blocked).
- **SC-004**: Restarting the app restores 100% of open tabs, unsaved edits, active tab, sidebar
  state and layout for every workspace.
- **SC-005**: Every request, folder and collection created in Xenon is a readable YAML file whose
  location mirrors the sidebar; a reviewer can find any request on disk by following its sidebar
  path, and a `git diff` of one edited request touches only that request's file.
- **SC-006**: Switching between two workspaces never shows any item, tab or state from the other
  (verified across 100 switches with distinct content).
- **SC-007**: Zero data loss across forced quits during saves (no truncated or empty files after
  repeated kill-during-save tests).
- **SC-008a**: Re-opening any History entry reproduces the sent request exactly (100% of fields
  match) for the 1,000 most recent sends per workspace.
- **SC-008**: All of the above behave identically on macOS, Windows and Linux.

## Assumptions

- Postman's current layout is the reference for region placement (the PRD's layout diagram was
  not available).
- Removing a workspace defaults to keeping its files; trashing the folder is an explicit choice.
- Unsaved new tabs and unsaved edits are restored after restart (Postman behaviour).
- Responses are kept only in memory while the app runs; saving responses as examples is not part
  of the MVP.
- Collection and folder headers apply to all requests beneath them, nearer definitions win
  (PRD: collections and folders have their own headers).
- `{{variables}}` are accepted in all fields but sent literally until stage 2 adds environments.
- History stores requests as written (with `{{variables}}` unresolved), so once stage 2 adds
  secrets no secret value can end up in history.
- Cookies (jar and response Cookies tab), Test Results, the console's content, drag and drop in
  the sidebar, the tab right-click menu, the remaining hotkeys, code snippets and the Variables
  panel are later stages (2, 4, 5) and out of scope here.
- The 5 MB threshold implements the PRD's "over a few MB".
- The default request timeout is 0 (no timeout), as in Postman.
- Until stage 2 adds secret variables, an auth value typed directly into a request (e.g., a
  Bearer token) is stored in that request's file as typed, exactly as Postman exports it;
  stage 2 lets such values move into the encrypted vault via `{{variables}}`. Local drafts and
  history are owner-only files on the machine.
- Workspace folders on network file systems work, but outside-change detection (FR-017) is only
  guaranteed on local disks.
