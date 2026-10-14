# Feature Specification: MCP Server — Full Xenon Access for AI Clients, Secrets Write-Only

**Feature Branch**: `xen-008`

**Created**: 2026-10-13

**Status**: Draft

**Input**: User description: "XEN-008 (GitHub issue #3) — Xenon starts an MCP server with the app so any AI client on
this machine (Claude Code, Claude Desktop, Cursor, …) can work with everything inside Xenon. Full feature parity with
the UI, with one deliberate exception: secrets are write-only for AI — an AI can create secret variables and send
requests that use them, but can never read or decrypt a secret value. Token-efficient, fast, small well-named tool set."

## Context

Xenon's MVP (XEN-001 … XEN-007) is done: workspaces, the request tree, sending, environments and the secrets vault,
scripts and console, history, documentation, import/export, snippets, cookies and git sync. Every feature is already
reachable from the UI through one Go entrypoint layer. This feature adds a second entrypoint for AI clients, speaking
the Model Context Protocol (MCP), so a developer can ask their AI assistant to browse, build, send and maintain API
requests in Xenon. The AI works on the same live state the human sees.

The product rule that shapes everything here comes from the constitution ("Secrets Never Leak"). Secret values must
never reach an AI's chat log. An AI may write and use secrets, but every value it reads back is masked as `•••••`.

Decisions already made by the product owner (2026-10-13, before specification):

- **Transport**: Streamable HTTP bound to the loopback interface only, on a fixed default port the user can change in
  Settings. A small `xenon mcp` stdio command proxies to the running app for clients that only speak stdio.
- **Access control**: there is no token. The server binds to loopback only and is protected against DNS rebinding
  (Host and Origin checks). Any process running as the same OS user can connect. This exposure is accepted, because
  such a process can already read the vault key from the OS keychain. An "AI access (MCP)" on/off switch in Settings defaults to ON.
- **Leak guard**: a final filter over every MCP result masks any known secret value, in any of its common encodings.
  Two residual risks are accepted and documented: a value that a script deliberately transforms (for example,
  reversed), and an AI that sends `{{secret}}` to a destination it controls and reads the value there. The second is
  inherent to letting the AI send requests that use secrets. "Confirm before send" environments and the AI client's
  approval of `send` mitigate it. Files written on the AI's behalf (a saved response body, an example from a response)
  are AI output and are written masked.
- **Approvals**: an AI send that uses an environment marked "Confirm before send" asks the human to Approve or Deny in
  the Xenon UI. Other destructive tools are flagged destructive in their tool metadata, so the AI client's own approval
  prompt covers them. Vault setup, unlock, reset, reveal and copy are never exposed to AI.
- **Approval bypass (added by the product owner mid-specification)**: Settings has a tickable "Dangerously skip
  approvals" option, OFF by default. When it is ticked, AI sends to "Confirm before send" environments go through with
  no prompt in Xenon. It never relaxes the secret restriction.

## Clarifications

### Session 2026-10-13

- Q: Tool naming scheme? → A: Resource tools with an `action` argument (for example `tree` with
  `action: list|get|search`). Read actions, changing actions and destroying actions live in separate tools, so each
  tool's read-only / destructive hint is accurate for every action it accepts. The plan collects deletions in one
  `delete` tool with a `target`, which keeps the hint accurate and saves tokens. The whole set stays at 25 tools or
  fewer.
- Q: Which workspace do AI tools act on? → A: The shared active workspace. Switching through MCP switches the Xenon
  window too.
- Q: Should the human see what the AI did? → A: Yes, as a full activity log of every MCP tool call.
- Q: Where does the activity log live? → A: Inside the existing sidebar History view (product owner's direction). A
  source filter (All / User / AI) sits next to Clear History, and AI entries are tagged "AI · <client name>".
- Q: Which AI tool calls become History entries? → A: Every tool call, reads included. AI sends are normal send rows
  tagged AI. Every other call is a compact action row.
- Q (product owner, added mid-clarification): History caps? → A: Configurable per source in Settings. One cap covers
  User entries (the human's sends) and one covers AI entries (AI sends and action rows together). Both default to
  1,000, each with a range of 100–20,000. Example: AI 1,000 and User 10,000.
- Q (product owner, added mid-clarification): How does Settings open? → A: As a main-area tab, not a modal dialog. The
  sections are Scripts, History, AI access (MCP), Git accounts and Shortcuts.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Connect an AI client and browse/send (Priority: P1)

Luis has Xenon open. In Settings → AI access (MCP) he sees that the server is running, its address, the stdio command,
and a ready-to-paste client config. He pastes the config into Claude Code. He then asks it to "find the Get user
request in Acme and send it": the AI lists the workspace's collections, drills into a folder, reads the request, sends
it with the active environment, and reports the status, headers, timing and body.

**Why this priority**: Without connection, browsing and sending, there is no feature. This is the minimum useful
slice.

**Independent Test**: Start the app with a workspace that has a collection, a folder and a request pointing at a mock
target. Connect an MCP client over HTTP, list the tools, browse to the request, send it, and check that the result
matches what the UI would show.

**Acceptance Scenarios**:

1. **Given** Xenon is running with AI access ON, **When** an MCP client connects to the shown address, **Then** the
   connection succeeds and it can list Xenon's tools.
2. **Given** a stdio-only client configured with the shown `xenon mcp` command, **When** Xenon is running, **Then** the
   client reaches the same tools through the bridge. **When** Xenon is not running, or AI access is off, **Then** the
   bridge exits with a one-line message naming the cause: "Xenon is not running" or "AI access is off".
3. **Given** a connected client, **When** it asks for the tree with no arguments, **Then** it gets only the top level
   (collections: id, name, child count), not the whole tree. **When** it asks for one folder, **Then** it gets that
   folder's direct children (id, kind, name, and method for requests).
4. **Given** a saved request, **When** the AI sends it, **Then** Xenon resolves variables with the active environment,
   runs the request's scripts, records history exactly as a UI send does, and returns status, status text, time,
   size, response headers and the body (truncated per FR-022), plus test results.
5. **Given** an ad-hoc request (method, URL, headers, body) that is not saved, **When** the AI sends it, **Then** it is
   sent with the active environment's variables and nothing is written to the tree.
6. **Given** AI access is switched OFF in Settings, **Then** the server stops within 2 seconds and new connections are
   refused. Switching it back ON starts it again without restarting Xenon.
7. **Given** a request from a non-loopback address, or with a Host/Origin that is not the loopback address, **Then**
   it is rejected and the server never handles it.

---

### User Story 2 - Secrets are write-only for AI (Priority: P1)

Luis asks the AI to "add an API token secret to Staging and call /me with it". The AI creates a secret variable
`apiToken` and sets its value (straight into the vault). It sends a request using `{{apiToken}}`. The request succeeds.
Nowhere in anything the AI receives does the token's plaintext appear: not in variable listings, the resolved URL or
headers, history, the console, a code snippet, an export, a test result, an error, or the response itself (even when
the target echoes the header back).

**Why this priority**: This is the one restriction that makes the feature safe to ship. It is co-equal with Story 1.

**Independent Test**: With a workspace whose vault holds a known secret, and a mock target that echoes every request
header and the body back, call every MCP tool that can return data. Assert the plaintext (raw, URL-encoded,
JSON-escaped and base64 basic-auth forms) never appears, and `•••••` appears where the value would be.

**Acceptance Scenarios**:

1. **Given** an environment, **When** the AI adds a variable marked secret with a value, **Then** the value is stored
   in the vault, the workspace file holds only the name and `secret: true`, and the response shows the value as
   `•••••`.
2. **Given** an existing secret, **When** the AI replaces its value, **Then** the vault holds the new value and the
   response shows `•••••`.
3. **Given** a request using `{{apiToken}}` in a header, **When** the AI sends it to an echo endpoint, **Then** the
   send uses the real value (the target receives it), and every place it would appear in the AI's result shows
   `•••••`: echoed response headers, body and test messages.
4. **Given** any variable listing, inspection or usage view, **Then** secret values show `•••••` and plain values
   show in full.
5. **Given** the AI asks for a code snippet with secrets included, or a Xenon export with secrets included, **Then**
   the tool refuses with a clear message that including secrets is a human-only action in the Xenon UI.
6. **Given** the AI tries to change a secret variable to plain (which would reveal its value), **Then** the change is
   refused and the variable is unchanged. Changing a plain variable to secret is allowed and moves its value into the
   vault.
7. **Given** a script that copies a secret into a plain variable, a test name or a `console.log`, **When** the AI then
   reads that variable, the test results or the console, **Then** the copied value shows `•••••`.
8. **Given** the vault is locked (passphrase mode, not yet unlocked), **When** the AI sends a request needing secrets,
   **Then** it gets the same "secrets unavailable" result the UI gets, telling it to ask the human to unlock the vault
   in Xenon. There is no tool to unlock it.
9. **Given** the MCP tool list, **Then** there is no tool or action that reveals, copies or decrypts a secret, an OAuth
   token, a git credential, or that sets up, unlocks or resets the vault.

---

### User Story 3 - Build and edit collections from the AI (Priority: P2)

Luis asks the AI to "create a Payments collection with an Orders folder and a POST /orders request with a JSON body".
The AI creates them, and they appear in the Xenon sidebar immediately. It then renames, moves, duplicates and deletes
items, edits a request's URL, params, headers, body, auth and scripts, saves an example from a response, and edits
collection and folder settings (auth, headers, scripts, variables).

**Why this priority**: Authoring is the second most valuable use after browse/send. Without it, the AI is read-and-send
only.

**Independent Test**: Through MCP, create a collection → folder → request, then edit, rename, move, duplicate and
delete. Assert the files on disk match what a UI-created item would produce, and that the UI received a tree-changed
event for each change.

**Acceptance Scenarios**:

1. **Given** a workspace, **When** the AI creates a collection, a folder inside it and a request inside the folder,
   **Then** the files are written exactly as a UI create writes them, and the open Xenon window shows the new items
   without a reload.
2. **Given** a request open in a UI tab with unsaved edits, **When** the AI saves a change to the same request,
   **Then** the UI treats it like an external change on disk (its existing "changed on disk" handling), and no edit is
   silently lost.
3. **Given** a delete, workspace remove, history clear, cookie clear or git push tool, **Then** its tool metadata marks
   it destructive. Every read-only tool is marked read-only.
4. **Given** collection or folder settings, **When** the AI updates auth, headers, scripts or variables, **Then** they
   are saved exactly as the UI's settings save does.
5. **Given** a collection, **When** the AI reads its documentation, **Then** it gets the OpenAPI YAML. **When** it
   applies an edited document, **Then** Xenon returns the same plan the UI shows (updated / created / deleted). Deletes
   are applied only when the AI explicitly confirms them in the apply call.

---

### User Story 4 - Environments, globals and the rest of Xenon (Priority: P2)

The AI lists environments, switches the active one (the switcher in the UI updates), reads and edits variables
(initial and current values), and uses history, import and export, code snippets, cookies and git sync (status,
changes, commit, pull, push, conflict resolution).

**Why this priority**: Completes parity. Each area is small once the entrypoint exists.

**Independent Test**: For each area, call its tool and compare the result with the same operation through the UI
entrypoint on the same fixture.

**Acceptance Scenarios**:

1. **Given** several environments, **When** the AI sets one active, **Then** the UI switcher shows it as active.
2. **Given** a variable, **When** the AI sets its current value, **Then** the current value changes and the initial
   value (in the workspace file) does not. **When** it sets the initial value, **Then** the file changes.
3. **Given** history entries, **When** the AI lists history, **Then** it gets compact rows, newest first, paged. Send rows
   carry id, time, source, method, URL as written, status and duration. Action rows carry id, time, client, tool,
   action, target name and outcome.
4. **Given** a Postman collection file path, **When** the AI imports it, **Then** the result matches a UI import of
   that file, including clash handling.
5. **Given** an item, **When** the AI exports it to a file path (Xenon YAML or Postman) without secrets, **Then** the
   file matches a UI export with "Include secrets" off.
6. **Given** a linked workspace, **When** the AI asks for sync status, changes, commit, pull or push, **Then** each
   behaves as in the Sync panel. Conflicts are resolved per item with the same choices the UI offers.
7. **Given** the AI asks for workspaces, **Then** it gets id, name and which one is active. **When** it switches
   workspace, **Then** the Xenon window switches too (one shared active workspace).

---

### User Story 5 - Approval for "Confirm before send" environments (Priority: P2)

Luis's Production environment has "Confirm before send". When the AI sends a request while Production is active,
Xenon comes to the front with an Approve / Deny prompt naming the AI client, the method, the URL as written and the
environment. On Approve the request is sent and the AI gets the response. On Deny, or after 120 seconds with no answer,
the AI gets a "denied by the user" result and nothing is sent. With "Dangerously skip approvals" ticked in Settings,
the prompt is skipped and the request is sent.

**Why this priority**: It protects production from an AI acting alone, using a flag the user already set.

**Independent Test**: With a confirm-before-send environment active, send through MCP, and drive the approval prompt
through the UI entrypoint: Approve → sent; Deny → not sent; no answer within the timeout → not sent. Then tick the
bypass setting and confirm the send goes through with no prompt.

**Acceptance Scenarios**:

1. **Given** a confirm-before-send environment is active and the bypass is OFF, **When** the AI sends, **Then** the
   Xenon window comes to the front with the prompt, and the AI's call waits for the answer.
2. **Given** the human approves, **Then** the request is sent once and the AI receives the result.
3. **Given** the human denies, or does not answer within 120 seconds, **Then** nothing is sent and the AI receives
   "denied by the user in Xenon" (or "approval timed out").
4. **Given** "Dangerously skip approvals" is ticked, **Then** AI sends to confirm-before-send environments are sent with
   no prompt. The Settings checkbox shows a warning that the AI then sends to protected environments unattended.
5. **Given** "Dangerously skip approvals" is ticked, **Then** secrets are still write-only, and every masking rule still
   applies.
6. **Given** several pending AI approvals, **Then** each is shown and answered on its own, in arrival order.

---

### User Story 6 - See everything the AI did, in History (Priority: P2)

Luis opens the History view after a session with Claude Code. Next to Clear History is a source filter: All / User /
AI. On "AI" he sees every call the AI made, newest first. Sends are ordinary history rows with an "AI · Claude Code"
tag, and he can reopen them like any send. Every other call is a compact action row, for example "AI · Claude Code —
tree.create request 'POST /orders' in Payments / Orders · ok · 12 ms" or "AI · Claude Code — vars.list Staging · ok".
Expanding an action row shows its arguments, masked.

**Why this priority**: The human stays in control of an agent acting inside their tool. Placing this in History reuses
a view they already know.

**Independent Test**: Make a mix of read, write, destructive and send calls through MCP, plus one UI send. Then list
history through the UI entrypoint with each filter value. Assert the rows, tags, order, masking and caps.

**Acceptance Scenarios**:

1. **Given** the AI makes any tool call while a workspace is active, **Then** one History entry is recorded in that
   workspace. It is a send row (tagged AI) for sends, and an action row for everything else, reads included.
2. **Given** an action row, **Then** it shows time, client name, tool and action, target (the item's name and path,
   when the call names an item), outcome (ok, the error code, denied or timed out) and duration. Its expanded view shows
   the call's arguments with secret-carrying fields redacted and the leak guard applied, capped at 2 KB.
3. **Given** the filter is set to User, **Then** only the human's sends show. On AI, only AI rows (sends and actions)
   show. On All, everything shows, merged newest first. The choice is remembered per workspace like the sidebar view.
4. **Given** a filter is set, **When** Luis clicks Clear History, **Then** only the entries the filter shows are
   cleared. On All, everything is cleared.
5. **Given** Settings → History caps of AI 1,000 and User 10,000, **When** the 1,001st AI entry is recorded, **Then**
   the oldest AI entry (send or action) is dropped and no User entry is dropped. **When** the 10,001st User send is
   recorded, **Then** the oldest User send is dropped and no AI entry is dropped.
8. **Given** Luis lowers a cap in Settings and saves, **Then** History immediately shows at most the new cap of entries
   for that source, and the dropped entries are gone from disk the next time that workspace's history is written.
   Raising a cap drops nothing.
6. **Given** an AI call made while no workspace is active (for example, listing workspaces on first run), **Then** no
   History entry is written. The call still appears in the app log.
7. **Given** the AI reads history through MCP, **Then** it can filter by source too, and action rows come back compact
   (no argument payloads unless it asks for one entry).

---

### User Story 7 - Settings opens as a tab (Priority: P3)

Luis clicks the Settings (gear) button, or presses Cmd+, (Ctrl+, on Windows and Linux). Instead of a modal dialog, a
**Settings** tab opens in the main area next to his request tabs. It has sections for Scripts (script timeout), History
(User cap, AI cap), AI access (MCP) (switch, port, status, skip approvals, stdio command, client configs) and Git
accounts, plus a Shortcuts reference. He can switch back to a request tab and return to Settings without losing his
unsaved edits.

**Why this priority**: The new AI and History settings make the dialog too crowded. A tab matches how environments
already open. The feature works without it, so it is P3.

**Independent Test**: Frontend tests open Settings from the button and the hotkey, and check that one Settings tab
exists, that it is focused on a second open, and that edits survive tab switches. They also check the close prompt when
there are unsaved edits, and that Save calls the settings service with every field.

**Acceptance Scenarios**:

1. **Given** no Settings tab is open, **When** Luis clicks the gear button or presses Cmd/Ctrl+, , **Then** a pinned
   Settings tab opens and is focused. **Given** it is already open, **Then** it is focused, and no second Settings tab
   opens.
2. **Given** unsaved edits in the Settings tab, **Then** the tab shows the dirty dot. Closing it asks Save / Don't
   Save / Cancel, exactly like an environment tab.
3. **Given** Luis clicks Save, **Then** every field is validated (script timeout 1–60 s, caps 100–20,000, port
   1024–65535) and saved together. Invalid fields show inline errors and nothing is saved. Changes to AI access and the
   port apply immediately (FR-004).
4. **Given** Git accounts, **Then** adding, editing and removing accounts work inside the tab as they did in the
   dialog. They act immediately and are not part of Save.
5. **Given** the Settings tab is open when Xenon quits, **Then** it is restored on the next start, like other tabs.
   Unsaved edits are not restored.
6. **Given** the old Settings dialog, **Then** it no longer exists. Every entry point opens the tab.

---

### Edge Cases

- **Port already in use**: the server does not start. Xenon keeps working. Settings shows "Port N is in use" and lets
  the user pick another port, which starts the server at once. The stdio bridge checks that the server it reached is
  Xenon. If another program answers on the port, the bridge reports "port N is in use by another program" and forwards
  nothing.
- **Port changed while clients are connected**: existing sessions on the old port end. Clients reconnect using the new
  address shown in Settings.
- **Xenon quits with clients connected**: sessions end. A pending approval counts as denied, so nothing is sent.
- **No active workspace** (first run, all removed): workspace-scoped tools return "No active workspace — create or open
  one". Workspace list, create and open still work.
- **Item id not found or from another workspace**: a not-found result. The workspace boundary is never crossed. Tools
  act only on the active workspace.
- **Binary or very large response body**: binary bodies are never inlined. The AI gets content type and size, and can
  save the body to a file path. Text bodies longer than the limit are truncated with the total masked size and an
  offset/length way to read further.
- **A secret value shorter than 4 characters**: it is still masked when it appears as a whole value. Masking inside
  free text applies to values of 4 or more characters, so a 1-character secret doesn't mask every letter of a
  response.
- **A secret replaced or deleted during the session**: values the guard has seen during this app session stay masked
  until Xenon quits, so an old value echoed in a cached response is still hidden.
- **Two AI clients at once**: both work. Writes are serialized the same way UI writes are. Each sees the other's
  changes on the next read.
- **AI edits a request the human is editing**: covered by Story 3, scenario 2.
- **AI switches workspace**: the switch call is recorded in the workspace that is active when the call finishes (the
  new one).
- **History is cleared while the AI is working**: later calls are recorded as usual. Clearing never fails an AI call.
- **Settings with no workspace open** (first run): the gear button and hotkey still work. With no tab strip, the
  Settings page fills the main area, and closing it returns to the welcome screen. Once a workspace opens, Settings
  opens as a tab.
- **OAuth 2.0 "Get new token"**: allowed. It opens the system browser for the human to log in, and the AI receives
  token metadata only (type, expiry, scopes), never the token.
- **Script that transforms a secret** (for example, reverses or hashes it) before logging or storing it: the
  transformed value is not recognized and may reach the AI. This risk is accepted by the product owner and documented
  in the Settings section's help text.
- **Request body that fails validation**: the tool returns the same validation error code and message as the UI
  binding, so the AI can correct the call.

## Requirements *(mandatory)*

### Functional Requirements

**Lifecycle and connection**

- **FR-001**: Xenon MUST start the MCP server when the app starts, if AI access is ON (default ON), and stop it when
  the app quits.
- **FR-002**: The server MUST listen on the loopback interface only (127.0.0.1), on a configurable port with a fixed
  default. It MUST reject any request whose Host is not the loopback address and port, or that carries an Origin that
  is not loopback (DNS-rebinding protection). Each rejection is logged as a warning with the reason, never with the
  request body. Requests larger than 4 MiB are rejected. It MUST NOT require a token.
- **FR-003**: The Settings tab MUST have an "AI access (MCP)" section with: the on/off switch; the port (editable, validated
  1024–65535); the status (Running at `http://127.0.0.1:<port>/mcp`, Off, or the error); a "Dangerously skip approvals"
  checkbox (default OFF) with a warning line; the stdio command; ready-to-paste client configs for HTTP and for stdio,
  each with a copy button; and a short note on the two accepted residual risks (transformed values, and sending to a
  destination the AI controls).
- **FR-004**: Turning the switch, or changing the port, MUST take effect immediately, without restarting Xenon.
- **FR-005**: A `xenon mcp` command MUST bridge stdio to the running app's HTTP server. It reads the port from Xenon's
  settings. When Xenon is not running or AI access is OFF, it exits non-zero with a one-line explanation.
- **FR-006**: MCP settings (switch, port, bypass) MUST be stored with the other app settings on this machine, never in
  workspace files, and MUST NOT be changeable through MCP.

**Parity through the existing behaviour**

- **FR-007**: The MCP entrypoint MUST call the same controllers the UI's bindings call, with the same validation and
  error codes. It MUST NOT add business logic. Anything MCP-specific (masking, approval, compaction, truncation) lives
  in the MCP entrypoint layer and its supporting services.
- **FR-008**: The AI MUST be able to list, create, open (an existing folder) and switch workspaces, rename them, and
  remove them (removal is marked destructive). There is one shared active workspace. Every workspace-scoped tool acts on
  it, and switching it through MCP switches the Xenon window too.
- **FR-009**: The AI MUST be able to browse the tree level by level, search it, read any collection, folder, request
  or example, and create, update, rename, move, duplicate and delete them (delete is marked destructive).
- **FR-010**: The AI MUST be able to send a saved request or an ad-hoc request with the active environment, with the
  same scripts, history, cookies and console behaviour as a UI send, and get status, headers, body, timing, size, test
  results and errors.
- **FR-011**: The AI MUST be able to list environments, read and edit the globals, create, update, duplicate and
  delete environments (delete is destructive), set the active one, and read and edit variables. Editing includes
  initial and current values, enabled, description, add and remove, plain→secret, and the secret value.
- **FR-012**: The AI MUST be able to read and edit collection and folder settings: auth, headers, scripts, variables
  and description.
- **FR-013**: The AI MUST be able to: list and read history and clear it (destructive); read, preview-plan and apply
  documentation (deletes only with explicit confirmation in the call); detect and import from file paths; export to a
  file path (Xenon YAML or Postman, without secrets); generate code snippets (masked); list, add, edit and delete
  cookies, and clear them (destructive); and use git sync. Git sync covers status, changes, link, unlink (destructive),
  clone as a workspace, sync settings, commit, pull, push (destructive), conflict resolution, and finishing or
  cancelling a pull.
- **FR-014**: The AI MUST be able to read the console (masked) and clear it, run the script checker and read the
  script reference, read and save app settings except MCP settings, read examples and create them from the last
  response, and get OAuth 2.0 token metadata, request a new token, refresh, choose which token is used, and clear it.
- **FR-015**: These UI-only features MUST NOT be exposed: vault status changes (setup passphrase, unlock, reset),
  reveal and copy of any secret or OAuth token, git account management (adding or removing accounts, tokens, SSH keys,
  GitHub sign-in, host-key confirmation), file and folder picker dialogs (the AI passes paths directly), window state,
  tab and layout UI state, and the clipboard.
- **FR-016**: Every change made through MCP MUST appear in the open Xenon UI without a reload. This covers the tree,
  request contents, environments, variables, the active environment, the active workspace, history, cookies and sync
  status.

**Secrets — write-only for AI**

- **FR-017**: The AI MUST be able to create a secret variable with a value, replace a secret's value, convert a plain
  variable to secret, and delete a secret variable. This applies in every scope that can hold secrets: globals,
  environments and collection variables. Folders hold no variables. Secret values MUST go only to the vault, as in the
  UI.
- **FR-018**: The MCP entrypoint MUST refuse, with a "human-only in Xenon" error: converting a secret to plain; a
  snippet with secrets included; a Xenon export with secrets included; and any reveal or copy.
- **FR-019**: Every MCP tool result (its returned content, including error messages) MUST pass through one final
  leak guard. The guard replaces each known secret value with `•••••`. Known values are the active workspace's
  secret variables and OAuth tokens, plus every secret value the guard has seen in this app session. Each value is
  matched in its raw, URL-encoded, JSON-escaped and base64 forms, including basic-auth `user:password`, with the
  longest match first. Values of 4 or more characters are masked inside text. Shorter values are masked only when they
  are a field's entire value.
- **FR-020**: Secret values MUST appear as `•••••` in variable listings and inspections, resolved previews, history,
  console, snippets, test names and messages, script errors, response headers and bodies, and error messages returned
  to the AI.
- **FR-021**: The leak guard MUST NOT change what is sent to the target, what the human's own actions write to disk, or
  what is shown to the human in the UI. It applies to what the AI receives and to files written on the AI's behalf from
  response content (`export response_body`, an example created from a response). Those are written masked, and a
  binary body that contains a known secret is refused.

**Token efficiency and speed**

- **FR-022**: Responses MUST be compact by default:
  - Lists return ids and names (plus the one or two fields needed to tell items apart, such as kind, method and active
    marker), paged at 100 rows with a cursor.
  - Reads return one item.
  - Text response bodies are capped at 16 KB (of the masked text), with `truncated`, the total masked size, and
    offset/length paging over the masked text to read
    further.
  - Binary bodies are not inlined.
  - Empty or default-valued fields are omitted.
- **FR-023**: Tools MUST be resource tools with an `action` argument. Names are short and lowercase:
  - the resource name (`tree`) for reads;
  - `<resource>_edit`, or a plain verb (`send`, `import`, `export`), for changes;
  - one `delete` tool (taking a `target`) and `git_push` for deletions and pushes.

  A tool never mixes read-only actions with changing ones, so its read-only / destructive / idempotent hints hold for
  every action it accepts. A changing tool is marked destructive when any of its actions can remove data: an item, a
  variable or its vault value, a file it overwrites, a list it replaces, or local changes. Changing a field's value in
  place is not removal.
  The set MUST stay at 25 tools or fewer. Each tool description MUST be one sentence.
- **FR-024**: Listing the tree MUST return one level by default. A full-depth listing happens only when explicitly
  requested.
- **FR-025**: Errors MUST be returned as tool errors carrying Xenon's error code and message (masked), never as
  protocol failures, so the AI can correct the call.

**Approvals**

- **FR-026**: When an AI send would use an environment marked "Confirm before send" and "Dangerously skip approvals" is
  OFF, Xenon MUST bring its window to the front and show an Approve / Deny prompt. The prompt shows the client name,
  method, URL as written (unresolved, so no secret appears) and environment name and colour. The request is sent only
  on Approve.
- **FR-027**: A pending approval MUST resolve as denied after 120 seconds with no answer, when Xenon quits, or when
  the requesting client's session ends; its prompt is then removed. The AI
  receives "denied by the user in Xenon" or "approval timed out", and nothing is sent.
- **FR-028**: When "Dangerously skip approvals" is ON, confirm-before-send sends from AI MUST proceed with no prompt.
  The setting MUST NOT affect any secret rule or any other behaviour.
- **FR-029**: Missing-secret and expired-token acknowledgements MAY be passed by the AI in the send call, with the same
  effect as in the UI (references are sent unresolved, or the expired token is sent as is). They are not human
  approvals.

**AI activity in History**

- **FR-031**: Every MCP tool call made while a workspace is active MUST be recorded in that workspace's History. Sends
  become ordinary send entries with source "ai" and the client name. Every other call becomes an action entry with:
  time, client name, tool, action, target item (id, name and path, when the call names one), outcome (ok, error code,
  denied or timed out), duration, and the call's arguments (capped at 2 KB). The app log records each call's client,
  tool, action, outcome and duration, never its arguments. Before the leak guard runs, the
  arguments that carry secrets are replaced with `•••••` by field name, whether or not the call succeeded:
  `secret_value`, `value` in variable add and set-current calls, `value` beside `secret: true`, and `passphrase`. A
  `send` that never went out (refused, denied, timed out or invalid) is recorded as an action row. History stays in the
  app-data folder and is never synced, as today.
- **FR-032**: Human sends MUST be recorded with source "user". Existing history entries with no source count as "user".
- **FR-033**: The History view MUST offer a source filter (All / User / AI) next to Clear History, remembered per
  workspace. AI rows show an "AI · <client name>" tag. Clear History MUST clear only the entries the current filter
  shows.
- **FR-034**: History MUST cap entries per source, oldest dropped first. The User cap covers the human's sends. The AI
  cap covers AI sends and action entries together. Both caps are app settings (Settings → History), default 1,000,
  range 100–20,000, and are validated like the other settings. Lowering a cap trims each workspace's history for that
  source the next time that workspace's history is written or read. Recording an entry is a single append and MUST NOT
  fail the tool call; a write failure is logged only.
- **FR-035**: The MCP history tool MUST list entries with an optional source filter, compact by default, and return
  one entry's full (masked) detail on request.

**Settings as a tab**

- **FR-036**: Settings MUST open as a single, pinned main-area tab (gear button and Cmd/Ctrl+,), never as a modal
  dialog. Opening it again focuses the existing tab.
- **FR-037**: The Settings tab MUST hold the sections Scripts, History, AI access (MCP), Git accounts and Shortcuts.
  Its form fields (script timeout, history caps, AI access, port, skip approvals) save together with one Save button.
  Git account actions act immediately, as today.
- **FR-038**: The Settings tab MUST follow environment-tab behaviour: a dirty dot, Save / Don't Save / Cancel on close,
  and restoration after a restart (without unsaved edits).

**Product documentation**

- **FR-030**: The PRD MUST gain an "AI access (MCP)" section recording this feature, its secret rule, the History
  activity log and these decisions, so the feature is in scope under the constitution's "Postman Parity, Nothing More" principle.

### Key Entities

- **History settings**: user cap (int, default 1,000, 100–20,000), AI cap (int, default 1,000, 100–20,000). App-level.
- **MCP settings**: AI access enabled (bool, default true), port (int, default fixed), skip approvals (bool, default
  false). App-level, per machine.
- **MCP server status**: off, running (address), or error (reason, such as port in use). Derived at runtime and shown
  in Settings.
- **History entry (extended)**: adds source (user or ai), client name (for ai), and kind (send or action). Action
  entries carry tool, action, target (id, name, path), outcome, duration and masked arguments. They have no
  request/response.
- **Approval request**: id, client name, method, URL as written, environment (id, name, colour), created-at, outcome
  (pending, approved, denied, timed out). In memory only.
- **Known secret set**: the in-memory set of secret values (and their encoded forms) the leak guard masks. It is never
  written anywhere.
- **Tool**: name, one-sentence description, input schema, output shape, hints (read-only, destructive, idempotent).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With Xenon running, a supported MCP client connects and lists the tools in under 1 second on loopback.
- **SC-002**: An AI can browse to, send and read the result of a saved request in at most 4 tool calls.
- **SC-003**: Items created through MCP appear in the open Xenon UI within 1 second, with no manual refresh.
- **SC-004**: Across the BDD leak suite (every tool that returns data × raw, URL-encoded, JSON-escaped and base64
  forms × variables, history, console, snippets, exports, tests, errors, echoed response headers and body), 0 results
  contain a secret's plaintext.
- **SC-005**: The full tool list, names plus descriptions plus input schemas, fits in 6,000 tokens or fewer. A top-level
  tree listing of a 1,000-request workspace stays under 2,000 tokens.
- **SC-006**: Non-send tool calls on a 1,000-request workspace answer within 200 ms. A send to a local target that
  answers instantly completes, end to end through MCP, within 100 ms.
- **SC-007**: 100% of destructive tools carry the destructive hint, and 100% of read-only tools carry the read-only
  hint.

## Assumptions

- AI clients run on the same machine as Xenon. Remote access is out of scope.
- The "Confirm before send" flag, missing-secret preflight and expired-token preflight already exist in the backend and
  are reused unchanged. Approval adds a human prompt in front of the existing acknowledgement.
- Git account management stays human-only, because it handles credentials and host-key trust. Git sync actions on an
  already-linked workspace use the stored account.
- There is no rate limiting. The server is local and single-user, and History's AI cap bounds the storage that AI
  calls can use. This exclusion is intentional.
- Only one Xenon instance runs at a time (the app is already single-instance), so one port is enough.
- Supported clients are any that speak MCP over Streamable HTTP or stdio (Claude Code, Claude Desktop, Cursor and similar).
