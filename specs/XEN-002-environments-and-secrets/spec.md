# Feature Specification: Environments, Variables and Secrets (Stage 2)

**Feature Branch**: `xen-002`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 2 — Environments, variables and secrets: per-workspace environments, collection variables and globals, {{variable}} resolution with Postman precedence and dynamic variables, environment switcher, Variables panel, and secrets kept out of files in an encrypted per-workspace vault. Source: docs/prd/xenon-mvp-prd.md; builds on XEN-001."

## Context

Stage 2 of 7. XEN-001 delivered workspaces, requests stored as YAML, tabs, sending and History,
with `{{variables}}` accepted but sent literally. This stage makes variables real: each workspace
gets its own environments, collections get variables, the workspace gets globals, and every
`{{name}}` in a request is resolved when sending. Secret values never touch the workspace files
(which are meant for git); they live in an encrypted vault on the machine. Scripts that read and
set variables (stage 4), Postman import of environments (stage 3), code snippets (stage 5) and
secret export (stage 6) build on this model.

## Clarifications

### Session 2026-10-09

- Q: Folders have variables in the PRD but not in its scope list or in Postman — include them?
  → A: No. Only collections have variables; folders keep auth, headers and (stage 4) scripts.
- Q: Postman dynamic variables (`{{$guid}}`, `{{$randomEmail}}`, …)? → A: Support the full
  Postman list.
- Q: OS keychain unavailable? → A: Fall back to a passphrase typed once per app start.
- Q: One value per variable, or Postman's initial + current values? → A: Initial + current, like
  Postman.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Switch environments and send (Priority: P1)

Luis keeps `local`, `dev`, `hml` and `prod` environments in the Acme workspace, each with
`base_url`, `tenant` and so on. He writes `{{base_url}}/users` once, picks an environment in the
switcher at the top right, and sends. Switching to another environment sends the same request
somewhere else.

**Why this priority**: This is the reason environments exist; without it, every company and
stage needs duplicated requests.

**Independent Test**: Create two environments with different `base_url`, send the same saved
request with each selected, and observe two different targets.

**Acceptance Scenarios**:

1. **Given** environments `dev` (`base_url = https://dev.acme.test`) and `hml`
   (`base_url = https://hml.acme.test`), **When** Luis selects `dev` and sends
   `GET {{base_url}}/users`, **Then** the request goes to `https://dev.acme.test/users`;
   **When** he selects `hml` and sends again, **Then** it goes to `https://hml.acme.test/users`.
2. **Given** "No environment" is selected, **When** he sends `{{base_url}}/users` and no
   collection variable or global defines `base_url`, **Then** `{{base_url}}` is sent literally
   (as Postman does) and is highlighted red in the URL beforehand.
3. **Given** variables are used in params, path variables, headers, every body mode and auth
   fields, **When** he sends, **Then** every occurrence is replaced by its resolved value.
4. **Given** the active environment, a collection variable and a global all define `tenant`,
   **When** he sends, **Then** the environment's value is used; without the environment's, the
   collection's; otherwise the global's.
5. **Given** Acme has `dev` active and Globex has `prod` active, **When** he switches workspaces,
   **Then** each workspace shows and uses its own active environment; after a restart both are
   still selected.
6. **Given** any environment list, **When** he presses Cmd/Ctrl+Shift+E, **Then** the switcher
   opens with keyboard focus so he can pick an environment by typing and Enter.

---

### User Story 2 - Manage environments, collection variables and globals (Priority: P1)

Luis creates, renames, duplicates and deletes environments, and edits their variables in a table
with name, initial value, current value, type (plain/secret), enabled and description. The
initial value is stored in the workspace files and shared through git; the current value is his
local override on this machine. He can Reset a current value back to the initial value, or
Persist it as the new initial value. The same table edits collection variables (in collection
settings) and workspace globals.

**Why this priority**: Without editing there is nothing to resolve; initial vs current lets the
team share defaults while each person keeps local values.

**Independent Test**: Create an environment, add variables, set local current values, restart,
and inspect the workspace files and the local values separately.

**Acceptance Scenarios**:

1. **Given** a workspace, **When** Luis creates environment `dev` and adds `base_url` with initial
   value `https://dev.acme.test`, **Then** the environment is stored as a readable file in the
   workspace's environments folder containing the name and initial value.
2. **Given** a variable with initial value `A` and no current value, **Then** its current value
   shows `A` (inherited) and `A` is used when sending; **When** he sets the current value to `B`,
   **Then** `B` is used, the workspace file still says `A`, and `B` is kept only on this machine.
3. **Given** current `B` and initial `A`, **When** he chooses Reset, **Then** the current value
   goes back to `A`; **When** he instead chooses Persist, **Then** the initial value in the file
   becomes `B`. Reset all and Persist all apply to every row of the table.
4. **Given** environment `hml`, **When** he duplicates it, **Then** a new environment
   "hml Copy" has the same rows, initial values, current values (including secret values) and
   settings, and he can rename it to `prod`.
5. **Given** a disabled variable row, **Then** it is ignored when resolving, as if absent.
6. **Given** a collection, **When** he opens its settings, **Then** a Variables tab edits the
   collection's variables with the same table; **and** the environment switcher menu has a
   "Globals" entry that edits the workspace's globals the same way.
7. **Given** he deletes an environment (with confirmation), **Then** its file and its local
   current values and secret values are removed; if it was active, "No environment" becomes
   active.
8. **Given** a teammate changed an environment file through git, **Then** the new initial values
   appear without restart, and Luis's local current overrides are kept.

---

### User Story 3 - Keep secrets out of files (Priority: P1)

Luis marks `api_key` and `password` as secret. Their values are never written to the workspace
folder — only the name and the fact that it is secret — so the folder can go into the company's
git repository. Values are kept in an encrypted vault for that workspace on his machine, unlocked
automatically through the OS keychain. Secrets show as dots with an eye to reveal and a copy
button, and resolve normally when sending.

**Why this priority**: The PRD's core promise for a git-friendly tool; leaking a token is the
worst failure mode.

**Independent Test**: Create secret variables, grep the workspace folder for their values (no
match), restart and send a request that uses them (works without typing anything).

**Acceptance Scenarios**:

1. **Given** a secret variable `api_key = s3cr3t`, **Then** no file in the workspace folder
   contains `s3cr3t`; the environment file lists `api_key` as secret with no value.
2. **Given** the app restarts, **When** Luis sends a request using `{{api_key}}`, **Then**
   `s3cr3t` is sent without him typing a password (the vault unlocks through the OS keychain).
3. **Given** a secret row, **Then** its value shows as dots everywhere it is displayed
   (environment and Globals tabs, Variables panel, URL hover); the eye reveals it for that row;
   copy puts the real value on the clipboard.
4. **Given** a plain variable, **When** Luis switches it to secret, **Then** its current value
   moves into the vault, its initial value is removed from the file, and he is told the old
   initial value may still exist in git history.
5. **Given** a secret, **When** he switches it to plain, **Then** he confirms that the value will
   be written to the workspace files, and the current value becomes the initial value.
6. **Given** a secret variable that exists in the files but has no value on this machine (new
   machine, cloned repo), **Then** it is marked "missing" wherever it is shown; **When** he sends
   a request that uses it, **Then** he is warned which secrets are missing and can Send anyway
   (the placeholder is sent literally) or Cancel.
7. **Given** two workspaces, **Then** each has its own vault; a secret from one is never
   resolvable or visible in the other.

---

### User Story 4 - See and fix variables from the request (Priority: P2)

While editing a request, `{{variables}}` are highlighted: green when they resolve, red when they
do not, and hovering shows the value (masked for secrets) and where it comes from. Typing `{{`
suggests variable names. The right side panel's Variables view lists every variable the current
request uses — including those in inherited auth and headers — with resolved value and source
scope, editable in place; unresolved ones can be created in one step.

**Why this priority**: Makes variables visible and fixable without opening editors; important
for comfort, but stories 1–3 already work without it.

**Independent Test**: Open a request using resolved, unresolved and secret variables and check
the highlighting, hovers, suggestions and the Variables panel, including editing from it.

**Acceptance Scenarios**:

1. **Given** URL `{{base_url}}/users/{{user_id}}` with only `base_url` defined, **Then**
   `{{base_url}}` is green and `{{user_id}}` red; hovering `{{base_url}}` shows its value and
   "Environment: dev".
2. **Given** a key/value table (params, headers, form fields) containing `{{tenant}}`, **Then** it
   is highlighted the same way.
3. **Given** Luis types `{{ba` in the URL, **Then** a suggestion list shows matching variable names
   with their scope; Enter inserts `{{base_url}}`.
4. **Given** the Variables panel for that request, **Then** it lists `base_url` (value, scope),
   `user_id` (unresolved) and any variables used by inherited auth or headers; editing a value
   there changes the current value in its scope; for `user_id` he can choose "Add to environment
   / collection / globals" and set a value.
5. **Given** a secret in the Variables panel, **Then** its value is masked with reveal and copy,
   as in story 3.

---

### User Story 5 - Make dangerous environments obvious (Priority: P2)

Each environment can have a neon colour; `prod` gets red by default, so the switcher glows red
when prod is active. An environment can require confirmation before every send. An eye button
next to the switcher opens the active environment's tab (the Globals tab when none is active).

**Why this priority**: Prevents sending to production by accident; matters once environments
exist.

**Independent Test**: Create `prod`, check its default colour and glow, enable "Confirm before
send", send a request and cancel at the prompt.

**Acceptance Scenarios**:

1. **Given** Luis creates an environment named `prod` (or any name containing "prod",
   case-insensitive), **Then** its colour defaults to neon red; other new environments default to
   no colour; he can pick neon green, cyan, magenta, yellow, red or none.
2. **Given** an environment with a colour is active, **Then** the switcher shows that colour and
   glows in it.
3. **Given** an environment with "Confirm before send" on, **When** Luis sends any request,
   **Then** he must confirm in a prompt naming the environment before anything is sent; Cancel
   sends nothing (and records nothing in History).
4. **Given** the eye button next to the switcher, **When** he clicks it, **Then** the active
   environment's tab opens pinned in the main area (or is focused if already open); with no
   active environment the Globals tab opens instead (XEN-015).

---

### User Story 6 - Postman dynamic variables (Priority: P3)

Imported and hand-written requests that use Postman's built-in dynamic variables —
`{{$guid}}`, `{{$timestamp}}`, `{{$isoTimestamp}}`, `{{$randomInt}}`, `{{$randomEmail}}` and the
rest of Postman's list — get generated values when sent, exactly as in Postman.

**Why this priority**: Needed for Postman parity with real collections; requests without them
already work.

**Independent Test**: Send a request using several dynamic variables to an echo target and check
each value's format.

**Acceptance Scenarios**:

1. **Given** a body containing `{{$guid}}`, `{{$timestamp}}`, `{{$isoTimestamp}}`,
   `{{$randomInt}}` and `{{$randomEmail}}`, **When** sent, **Then** each is replaced by a value of
   the documented Postman format (UUID v4, Unix seconds, ISO-8601 UTC, integer 0–1000, email).
2. **Given** `{{$guid}}` appears twice in one request, **Then** each occurrence gets its own value
   (Postman behaviour).
3. **Given** a user-defined variable named `$guid` in any scope, **Then** the user's variable wins
   over the dynamic one (Postman's resolver checks dynamic variables last).
4. **Given** `{{$notARealOne}}`, **Then** it is treated as an unresolved variable (red, sent
   literally).
5. **Given** dynamic variables in the URL or a table, **Then** they are highlighted as resolved
   with a hover describing what they generate (no value is shown, since it is generated at send
   time).

---

### Edge Cases

- **Nested references**: a variable whose value contains `{{other}}` is resolved recursively, up
  to 19 passes (Postman's limit); a cycle (`a → b → a`) stops there and the remaining `{{…}}`
  text is sent literally and highlighted red.
- **Variable names**: case-sensitive and matched exactly, as in Postman — `{{ name }}` looks up
  " name " (with spaces); any characters except `{`, `}` and line breaks; the editor warns about
  leading/trailing spaces in a key; duplicate names in one table — the last enabled row wins and
  the editor flags the duplicates.
- **Keychain unavailable** (Linux without a Secret Service, access denied, locked keychain): Xenon
  asks for a passphrase once per app start to unlock vaults; the first time, it asks him to
  choose one (entered twice). Without the passphrase, secrets are "locked": shown as locked,
  resolution of secrets is blocked, and sending a request that uses one warns like a missing
  secret.
- **Forgotten passphrase or lost keychain entry** (new OS user, keychain reset): the vault cannot
  be opened; Xenon explains this and offers "Reset vault", which discards the stored secret values
  (all secrets become missing) after confirmation. Nothing is ever decrypted without the key.
- **Vault file corrupted or deleted**: treated like a lost key — secrets become missing, Reset
  vault recreates it.
- **Environment file malformed or hand-edited**: shown with an error marker in the environment
  list; other environments work (same rule as XEN-001 FR-016).
- **Environment file deleted by git while active**: "No environment" becomes active and Luis is
  told.
- **A secret value typed into a plain field** (e.g., a token pasted straight into a header):
  stored as typed — Xenon cannot know it is secret; the Variables panel offers "Move to secret
  variable" for any literal value selected in a field.
- **Very long values** (e.g., a 10 KB JWT): stored and resolved intact; tables and hovers truncate
  the display with an ellipsis.
- **Same variable name in environment and as a dynamic variable** (`$guid`): see story 6 #3.
- **Disabled current/initial**: a row's enabled flag is shared (in the file); disabled rows never
  resolve regardless of current value.
- **A teammate turns a plain variable into a secret through git**: the local plain current value
  for that row is moved into the vault on load (or deleted if the vault is locked/broken), so it
  never stays unencrypted on disk. The reverse (secret → plain) leaves the vault value unused;
  it is pruned on the next vault save.
- **Empty current value**: setting a current value to an empty string clears the override (the
  initial value applies), matching Postman's fallback; to send an empty value, set the initial
  value to empty.
- **Changing the passphrase** is not part of the MVP; "Reset vault" is the way out of a
  forgotten passphrase.
- **Workspace removed**: its vault, keychain entry and local current values are deleted with it
  (also when the folder is kept on disk).

## Requirements *(mandatory)*

### Functional Requirements

**Environments**

- **FR-001**: Each workspace MUST support any number of environments with free, unique (per
  workspace, case-insensitive) names, created, renamed, duplicated and deleted from the
  environment switcher menu and an environments manager.
- **FR-002**: Each environment MUST be stored as its own file in the workspace's `environments/`
  folder containing its id, name, colour, "Confirm before send" flag and variable rows (initial
  values only; never secret values; never current values).
- **FR-003**: Duplicating an environment MUST copy rows, initial values, settings, local current
  values and secret values into a new environment named "<name> Copy".
- **FR-004**: Deleting an environment MUST remove its file, its local current values and its
  secret values; if active, "No environment" becomes active.
- **FR-005**: Each environment MUST have a colour from {none, neon green, cyan, magenta, yellow,
  red}; a new environment whose name contains "prod" (case-insensitive) MUST default to red,
  others to none.
- **FR-006**: Each environment MUST have a "Confirm before send" flag (default off; default on
  for names containing "prod"). When the active environment has it on, every send MUST first ask
  for confirmation naming the environment; cancelling sends nothing and records no History entry.

**Variables and scopes**

- **FR-007**: Variable rows (environment, collection, globals) MUST have: name, initial value,
  current value, type (plain | secret), enabled, description.
- **FR-008**: Initial values of plain variables MUST be stored in the workspace files. Current
  values MUST be stored only on this machine (local app data, per workspace), never in the
  workspace folder. A row without a current value MUST use its initial value.
- **FR-009**: Users MUST be able to Reset (current := initial) and Persist (initial := current)
  each row, and Reset all / Persist all for a table. Persist MUST NOT be available for secret
  rows (they have no initial value).
- **FR-010**: Collections MUST have variables (edited in collection settings); folders MUST NOT.
  Each workspace MUST have globals, edited from a "Globals" entry in the environment switcher
  menu and stored in `workspace.yaml`.
- **FR-011**: Resolution precedence MUST be, strongest first: request-local (set by scripts,
  stage 4) → active environment → collection of the request → workspace globals. Disabled rows
  MUST be ignored.
- **FR-012**: `{{name}}` MUST be resolved when sending in: URL, query params, path variables,
  headers (keys and values), every body mode (raw content, form-data keys/values/text, urlencoded,
  GraphQL query and variables, binary file path), auth fields, and collection/folder inherited
  auth and headers. Resolution MUST happen in the backend at send time on a copy; the stored
  request and its History entry keep the unresolved text.
- **FR-013**: Nested references MUST resolve recursively up to 19 passes (Postman's limit); cycles
  and unresolvable names MUST be sent literally as `{{name}}`.
- **FR-014**: Unresolved variables MUST NOT block sending (Postman behaviour), except missing or
  locked secrets (FR-022).
- **FR-015**: Each workspace MUST remember its active environment (local, per machine), restored
  on switch and restart.

**Dynamic variables**

- **FR-016**: All Postman dynamic variables (`{{$guid}}`, `{{$timestamp}}`, `{{$isoTimestamp}}`,
  `{{$randomInt}}`, `{{$randomUUID}}` and the full `$random*` list in Postman's documentation —
  118 names) MUST resolve at send time to values of Postman's documented format, with a fresh
  value per occurrence, checked after all user scopes (a user variable of the same name wins).
  The list is published in contracts/dynamic-variables.md. Unknown `{{$…}}` names are unresolved.

**Secrets**

- **FR-017**: A secret row MUST store no value in any workspace file — only its name, `secret:
  true`, enabled and description. Its value (current value only) MUST be stored in an encrypted
  vault file for that workspace in local app data.
- **FR-018**: The vault's key MUST be stored in the OS keychain (macOS Keychain, Windows
  Credential Manager, Linux Secret Service) under one entry per workspace, so no password is
  needed in normal use.
- **FR-019**: When the keychain is unavailable, Xenon MUST fall back to a passphrase: chosen once
  (entered twice) when the first vault needs a key, then asked once per app start when a vault
  must be unlocked. The passphrase is never stored. While locked, secrets are shown as locked and
  treated as missing for sending.
- **FR-020**: Secret values MUST be masked (`•••••`) wherever displayed — environment and Globals
  tabs, Variables panel, hovers — with a per-row reveal and a copy button that copies the
  real value. (Console and code-snippet masking arrive with stages 4–5 and MUST use the same
  rule.)
- **FR-021**: Changing a row from plain to secret MUST move its current value into the vault,
  remove its initial value from the file, and warn that git history may still contain the old
  value. Changing from secret to plain MUST require confirmation and write the value as the
  initial value.
- **FR-022**: A secret defined in the files but without a value on this machine MUST be marked
  "missing". Sending a request that uses missing or locked secrets MUST first show which ones,
  with Send anyway (placeholders sent literally) or Cancel.
- **FR-023**: When the vault cannot be decrypted (lost key, forgotten passphrase, corrupted
  file), Xenon MUST explain it and offer "Reset vault", which discards stored secret values after
  confirmation. Xenon MUST never write secret values unencrypted anywhere (files, logs, History,
  UI state, temp files). Text Xenon itself produces from a resolved request — error messages and
  hints, History result summaries, log lines — MUST replace every occurrence of a secret value
  with `•••••`. (A response body that echoes a secret is server content and is shown as
  received; it is only ever kept in memory or in owner-only temp files.)
- **FR-024**: Removing a workspace MUST delete its vault, keychain entry and local current
  values.

**Editing experience**

- **FR-025**: `{{name}}` in the URL and in key/value tables MUST be highlighted green when
  resolvable, red when not, with a hover showing the value (masked for secrets, "generated at
  send" for dynamic variables) and its source scope.
- **FR-026**: Typing `{{` in the URL, key/value tables and body editors MUST offer suggestions of
  matching variable names (all scopes plus dynamic variables) with their scope.
- **FR-027**: The right side panel MUST have a Variables view listing every variable used by the
  current request (including inherited auth/headers), with resolved value, source scope and
  status (resolved, unresolved, missing secret, locked); values MUST be editable in place
  (current value in the source scope); unresolved variables MUST offer "Add to environment /
  collection / globals".
- **FR-028**: An environment switcher MUST always be visible at the top right with the active
  environment (colour + glow) and a "No environment" option; Cmd/Ctrl+Shift+E MUST open it with
  focus. An eye button next to it MUST open the active environment's tab
  (the Globals tab when none is active).
- **FR-029**: Changes to environment files and `workspace.yaml` made outside Xenon MUST be picked
  up live (XEN-001 FR-017 behaviour), keeping local current values and secrets.

### Key Entities

- **Environment**: id, name, colour, confirm-before-send, variable rows; belongs to one
  workspace; stored as one file.
- **Variable row**: name, initial value (plain only, in files), type, enabled, description; plus
  a local **current value** (plain: local state; secret: vault).
- **Collection variables / Globals**: variable rows owned by a collection (in its settings file)
  or by the workspace (in `workspace.yaml`).
- **Local variable state** (per workspace, per machine): current values of plain rows for every
  environment, collection and globals; the active environment id.
- **Vault** (per workspace, per machine): encrypted map of secret current values, keyed by scope
  and variable; unlock mode (keychain or passphrase).
- **Resolution result**: for a request, every referenced name with resolved value, source scope
  and status — used for sending, highlighting and the Variables panel.
- **Dynamic variable**: a `$`-prefixed name with a generator producing Postman-format values.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Switching the active environment and re-sending takes at most 2 interactions
  (pick environment, Send/Cmd+Enter).
- **SC-002**: After creating secrets in a workspace, a byte search of the whole workspace folder
  for each secret value finds 0 matches; the same holds for logs, History and UI-state files.
- **SC-003**: After an app restart with the keychain available, 100% of requests using secrets
  send correctly without the user typing anything.
- **SC-004**: For requests referencing up to 200 variable occurrences, highlighting and the
  Variables panel update within 100 ms of a keystroke and resolution adds under 10 ms to a send.
- **SC-005**: Every Postman dynamic variable in the published list produces a value matching its
  documented format in 1,000 consecutive generations.
- **SC-006**: Teammates sharing a workspace through git see identical initial values while their
  local current values and secrets stay private (0 local values in the folder).
- **SC-007**: With "Confirm before send" on, 0 requests are sent without an explicit confirmation.
- **SC-008**: All behaviour is identical on macOS, Windows and Linux, including the keychain and
  passphrase paths.

## Assumptions

- Postman semantics are the reference for precedence, initial/current values, disabled rows,
  dynamic variables and unresolved variables (sent literally).
- The "data" scope (collection runner data files) is not in the MVP; the runner is a PRD
  non-goal.
- Environment colour and "Confirm before send" live in the environment file, so they are shared
  with teammates through git; the active environment is per machine.
- One passphrase covers all workspaces' vaults when the keychain is unavailable.
- Duplicate names across scopes are normal (that is how precedence works); duplicates inside one
  table are flagged but allowed.
- Scripts setting variables (stage 4), Postman environment import (stage 3), code-snippet masking
  (stage 5) and encrypted secret export (stage 6) are later stages; this stage provides the
  scopes, the resolver and the vault they use.
