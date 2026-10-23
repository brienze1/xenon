# Feature Specification: Alias Environments

**Feature Branch**: `xen-018`

**Created**: 2026-10-23

**Status**: Draft

**Input**: User description: "Link request aliases to environments: an alias can be limited to one or more environments or allowed in all. An alias linked to environments is usable only while one of them is active; the backend refuses to send it otherwise, from the app and from AI clients. Source: https://github.com/brienze1/xenon/issues/25"

## Context

Request aliases (XEN-016) are named, independent versions of a request, stored with the request
in its file and picked from the request tab's **Aliases** dropdown. Every alias can be picked and
sent whatever environment is active. Nothing ties an alias to the environment it was made for:

- **An HML-only test.** An alias that tests something in HML (staging) can still be sent while
  **Prod** is active, with body values and headers never meant for Prod.
- **Logins.** A login alias carries a user and password that only exist in one environment.
  `mx-test-user` only exists in HML; sending it to Prod is wrong at best and leaks a test
  credential at worst.

Today the only safeguard is remembering which alias belongs where. This feature gives each alias
an **Environments** setting: **All environments** (the default, and what every existing alias
has) or a chosen set of the workspace's environments.

### Decisions taken before this spec (user, 2026-10-23)

| Question (issue #25) | Decision |
| --- | --- |
| Aliases for other environments in the dropdown | **Shown greyed out**, with their environments, and not selectable. Their ⋯ menu (change environments, delete) still works. |
| Active environment changes while a tab shows an alias that is no longer available | **The tab keeps the alias** (its unsaved edits are kept), **Send is blocked** and a message names the environments the alias is for. The backend refuses the send anyway. |
| No active environment | **Linked aliases are unavailable.** "No environment" is not one of the alias's environments. |
| One environment or several | **A set of one or more environments** (for example HML and DEV but not Prod), or All environments. |
| A linked environment is deleted (in the app or by a git pull) | **The alias stays linked and the leftover is flagged.** The stored id is kept, so the alias fails safe: it is never widened to All environments. If none of its environments exists any more it is unavailable everywhere. The dropdown and the Environments dialog show the leftover as **Deleted environment** so the user can relink it. |
| Postman export | The exported alias request's description **gets a note** naming the alias's environments (for example `Xenon alias for environments: HML`). Postman has no way to enforce it. |
| History | **Every send's history entry records the environment that was active** (alias or not), as it was at send time, and the History list shows it. |

### Settled from the issue and the code rather than asked

- **Link by environment id.** Every environment file has a stable `id` that a rename keeps
  (`internal/domain/entity/environment.go`), so the link survives renames.
- **Git and teams.** Environments are files in the workspace's `environments/` folder and request
  files live in the same workspace, so both are synced through git together; teammates working on
  the same workspace share the same environment ids. Environments created separately by hand in
  another workspace are different environments, and a link to them is not expected to carry over.
- **The Default can't be linked.** The issue's examples keep the Default as the environment-free
  version ("Default uses no fixed user") and its scope only adds the link to aliases. The Default
  is always available.
- **Where the guard lives.** Every send of a saved request with an alias, from the app or from the
  MCP `send` tool, goes through the one backend send (`RequestService.Send` /
  `internal/application/service/send_request.go`), which already resolves the alias (`ALS-01404`)
  and the active environment. The guard is placed there, so every current and future caller (for
  example the collection runner, XEN-017, still unmerged) inherits it without its own rule.
- **Scripts.** Scripts can't pick an alias of another request (nested sends carry no alias), so
  there is no script path to guard.
- **Existing files.** An alias without the setting means All environments; files written before
  this feature load and behave unchanged, and an alias set to All environments is written without
  the setting so its file does not change.

## Clarifications

### Session 2026-10-23

- Q: How does the History list show the environment active for each send? → A: The row shows the environment's colour dot and name as they were at send time, after the request/alias label; a send with no environment shows nothing; AI clients' history reads include the environment too.
- Q: Does the "Send blocked" message offer a one-click action? → A: Yes — one **Switch to <environment>** button per linked environment that still exists; it sets the workspace's active environment exactly as the top-right switcher does.
- Settled from the code (not asked): duplicating a request copies its file, so its aliases keep their links; duplicating an environment creates a new id (`Environment.Duplicated`, `internal/domain/entity/environment.go`), so aliases linked to the original are not linked to the copy.
- Settled from the code (not asked): changing an alias's environments is its own alias write (like Delete) that saves immediately; it never saves or discards the alias's unsaved content edits.
- Settled from the code (not asked): AI clients name environments by id when setting the link, as every other MCP tool does (ids from the `env` tool), and read back ids with names.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Limit an alias to environments and be stopped from sending it elsewhere (Priority: P1)

Luis's **Login** request has the alias `mx-test-user`, whose user only exists in HML. He opens the
Aliases dropdown, opens the ⋯ menu next to `mx-test-user` and picks **Environments…**. A dialog
lists **All environments** and every environment of the workspace, each with its colour dot. He
picks **HML** only and saves. The dropdown now shows `mx-test-user` with HML's colour dot and
name. With HML active he sends it as before. He switches the active environment to Prod: the tab
keeps showing `mx-test-user`, Send is disabled and a message says the alias is only for HML. Even
if a send reaches the backend (for example from an AI client), it is refused with a clear error
and nothing is sent.

**Why this priority**: This is the safety the feature exists for: a credential or test meant for
one environment never reaches another.

**Independent Test**: Link an alias to one environment, then send it with that environment
active (sent), with another environment active (refused, nothing sent) and with no environment
active (refused, nothing sent).

**Acceptance Scenarios**:

1. **Given** alias `mx-test-user` linked to HML, **When** it is sent with HML active, **Then** the
   request is sent as today.
2. **Given** alias `mx-test-user` linked to HML, **When** it is sent with Prod active, **Then**
   the send is refused with an error that names the alias's environments, nothing is sent, and no
   history entry is written.
3. **Given** alias `mx-test-user` linked to HML, **When** it is sent with no environment active,
   **Then** the send is refused the same way.
4. **Given** alias `qa-smoke` linked to HML and DEV, **When** it is sent with DEV active, **Then**
   it is sent; with Prod active it is refused.
5. **Given** an alias set to All environments, **When** it is sent with any environment or none
   active, **Then** it is sent as today.
6. **Given** any linked aliases on the request, **When** the Default is sent with any environment
   or none active, **Then** it is sent as today.
7. **Given** a tab showing `mx-test-user` (linked to HML) with HML active, **When** the active
   environment switches to Prod, **Then** the tab keeps showing the alias with its unsaved edits,
   Send is disabled and a message says the alias is only for HML with a **Switch to HML** button;
   pressing it (or switching back to HML any other way) makes HML active and enables Send again.

---

### User Story 2 - See which environment each alias belongs to (Priority: P1)

Luis opens the Aliases dropdown with HML active. `mx-test-user` (HML) is selectable and shows
HML's colour dot and name; `prod-smoke-user` (Prod) is listed greyed out with Prod's colour dot
and name and can't be picked; `v2` (All environments) shows no environment and is selectable, as
is the Default.

**Why this priority**: The dropdown is where aliases are picked; without the marking the user
can't tell why an alias is unavailable or which one to use.

**Independent Test**: With linked and unlinked aliases on a request, open the dropdown under
different active environments and check which entries are selectable and what each shows.

**Acceptance Scenarios**:

1. **Given** aliases linked to HML, to Prod, and to All environments, **When** the dropdown is
   opened with HML active, **Then** the HML and All environments aliases and the Default are
   selectable, and the Prod alias is listed greyed out and not selectable.
2. **Given** a linked alias, **When** the dropdown lists it, **Then** each of its environments is
   shown with the environment's current colour dot and name (a renamed environment shows its new
   name).
3. **Given** a greyed-out alias, **When** its ⋯ menu is opened, **Then** **Environments…** and
   **Delete** both work.
4. **Given** no environment is active, **When** the dropdown is opened, **Then** every linked
   alias is greyed out and the All environments aliases and the Default are selectable.

---

### User Story 3 - Choose environments when creating an alias (Priority: P2)

Luis picks **Create new alias…**. The dialog asks for the name and, below it, the alias's
environments, preset to **All environments**. He types `canary`, picks HML and creates it: the
new alias is linked to HML from the start.

**Why this priority**: Saves a second step, but linking afterwards from the ⋯ menu covers the
same need.

**Independent Test**: Create an alias with a chosen environment and with the preset, then read
both back.

**Acceptance Scenarios**:

1. **Given** the Create New Alias dialog, **When** it opens, **Then** the environments setting
   shows All environments.
2. **Given** the dialog, **When** the user creates `canary` with HML picked, **Then** the new
   alias is stored linked to HML.
3. **Given** the dialog, **When** the user creates an alias without changing the setting,
   **Then** the alias is stored as All environments, exactly as aliases are created today.
4. **Given** the user creates a linked alias while another environment is active, **Then** the
   alias is created and shown in the tab, Send is blocked with the message from User Story 1, and
   it is greyed out in the dropdown.

---

### User Story 4 - AI clients read, set and respect the link (Priority: P2)

An AI client reads the Login request through MCP and sees, for each alias, its environments (ids
and names) or that it is for all environments. It creates or edits an alias with an environments
list, or clears it back to all environments. When it sends `mx-test-user` while Prod is active,
the send is refused with the same error the app gets, and nothing is sent.

**Why this priority**: AI clients can already send aliases (XEN-016); without the guard they would
bypass the safety this feature adds.

**Independent Test**: Through MCP, read a request with linked and unlinked aliases, set and clear
a link, and send a linked alias with another environment active.

**Acceptance Scenarios**:

1. **Given** a request with linked and unlinked aliases, **When** an AI client reads it, **Then**
   each alias reports its environments (id and name), or none meaning all environments.
2. **Given** an alias, **When** an AI client edits it with an environments list, **Then** the link
   is stored; with an empty list, the alias becomes All environments.
3. **Given** an AI client creates an alias with an environments list, **Then** the alias is
   created linked to those environments.
4. **Given** an alias linked to HML, **When** an AI client sends it with Prod active, **Then** the
   send is refused with the same error code as in the app, nothing is sent and no approval prompt
   is shown.

---

### Edge Cases

- **Linked environment deleted.** The alias keeps the id; it is unavailable unless another of its
  environments is active; the dropdown and dialog show **Deleted environment** (FR-012).
- **All of an alias's environments deleted.** The alias is unavailable everywhere until relinked
  or set to All environments; the Send-blocked message says its environments were deleted and
  points to the Aliases menu, with no Switch button.
- **Environment renamed or recoloured.** The link is by id, so the alias keeps its link and shows
  the new name and colour.
- **Duplicate or unknown ids in an edit.** An edit naming the same environment twice stores it
  once; an edit naming an id that is not an environment of the workspace is refused (FR-011).
- **Active environment that no longer exists.** The backend already resolves sends without an
  environment when the remembered active environment was deleted; for the guard this counts as no
  active environment.
- **"Confirm before send" environments.** The alias guard runs before the confirmation step, so a
  refused alias never asks for confirmation.
- **Alias on screen with unsaved edits when switching environments.** Edits stay; only Send is
  blocked (User Story 1, scenario 7).
- **Concurrent edit of the request file.** Changing an alias's environments is an alias write and
  uses the same stale-write protection as the other alias writes.
- **Duplicating.** A duplicated request keeps its aliases' links; a duplicated environment is a
  new environment that no alias is linked to until the user adds it.
- **Changing environments with unsaved alias edits.** The environments change is saved on its own;
  the unsaved content edits stay as a draft.
- **Save As… on an alias.** Creates a new request whose Default is the version on screen; the
  Default can't be linked, so the link is not copied.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each alias MUST have an environments setting that is either All environments or a
  set of one or more of the workspace's environments, identified by their stable ids.
- **FR-002**: The setting MUST be stored with the alias in the request's file. An alias set to All
  environments MUST be stored without the setting, and an alias without the setting MUST be read
  as All environments, so existing files load and save unchanged.
- **FR-003**: An alias MUST be available when it is set to All environments, or when an
  environment is active and that environment is in the alias's set. It MUST be unavailable
  otherwise, including when no environment is active.
- **FR-004**: The Default MUST always be available and MUST have no environments setting.
- **FR-005**: The backend send MUST refuse an unavailable alias with a new, distinct alias error code,
  whose message names the alias's environments, before any script runs, any
  confirmation is asked or anything is sent; nothing MUST be sent and no history entry written.
- **FR-006**: FR-005 MUST apply to every send of a saved request with an alias, from the app and
  from AI clients over MCP; an AI send refused this way MUST NOT show an approval prompt.
- **FR-007**: Users MUST be able to change an alias's environments from an **Environments…** item
  in the alias's ⋯ menu in the Aliases dropdown, which opens a dialog offering All environments or
  any set of the workspace's environments (each with its colour dot and name).
- **FR-008**: The Create New Alias dialog MUST offer the same environments choice, preset to All
  environments.
- **FR-009**: The Aliases dropdown MUST show each linked alias's environments (colour dot and
  current name), MUST list unavailable aliases greyed out and not selectable, and MUST keep their
  ⋯ menu working.
- **FR-010**: When the alias on screen is or becomes unavailable (the active environment changes,
  or the alias is linked elsewhere), the tab MUST keep showing it with its unsaved edits, MUST
  disable Send and MUST show a message naming the alias's environments, with one **Switch to
  <environment>** button per linked environment that still exists (pressing it sets the active
  environment as the top-right switcher does); Send MUST be enabled again as soon as the alias
  becomes available.
- **FR-011**: Setting an alias's environments MUST refuse an id that is not an environment of the
  workspace with a distinct error code and MUST store each environment once.
- **FR-012**: When a linked environment is deleted, the alias's stored link MUST stay as it is
  (deleting an environment never rewrites request files). An id that matches no environment of the
  workspace MUST never make the alias available, MUST NOT widen it to All environments, and MUST
  be shown in the dropdown and the Environments dialog as **Deleted environment**, which the user
  can remove by saving a new choice.
- **FR-013**: AI clients MUST see each alias's environments (ids and names, or none for All
  environments) when reading a request's aliases, and MUST be able to set and clear them when
  creating or editing an alias.
- **FR-014**: The other existing alias behaviour (XEN-016) MUST keep working unchanged for linked
  and unlinked aliases.
- **FR-015**: A Postman export MUST add a note naming a linked alias's environments to the
  description of that alias's exported request; aliases set to All environments export unchanged.
- **FR-016**: Every history entry of a send (app or AI, alias or Default) MUST record the
  environment that was active at send time — its id, name and colour as they were then — or that
  none was; the History list MUST show its colour dot and name on the entry's row, after the
  request/alias label, and AI clients' history reads MUST include it. Entries with no environment,
  and entries written before this feature, show no environment.

### Key Entities

- **Request alias** (existing, XEN-016): a named, independent version of a request. Gains an
  **environments** setting: empty for All environments, or the ids of one or more environments.
- **Environment** (existing): a workspace file with a stable id, a name and a colour; the active
  environment is chosen per workspace.
- **History entry** (existing): gains the environment active at send time (id, name, colour).
- **Alias availability** (rule, not stored): whether an alias may be picked and sent given the
  active environment (FR-003).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of sends of a linked alias with any other environment active, or none, are
  refused with nothing sent, from both the app and AI clients.
- **SC-002**: 100% of existing request files (aliases without the setting) load, send and save
  exactly as before, with no change to their file contents.
- **SC-003**: A user can link an existing alias to an environment in under 15 seconds from the
  Aliases dropdown, and when creating an alias with no extra dialog.
- **SC-004**: In the Aliases dropdown a user can tell, without opening anything else, which
  environments each alias belongs to and which aliases are usable right now.
- **SC-005**: Renaming a linked environment never breaks or changes any alias's link.

## Assumptions

- The workspace's environment list and the active environment are already available to the
  request tab (the top-right environment switcher), so the dropdown can react to switches without
  extra user steps.
- The collection runner (XEN-017) is not on `main`; it inherits the guard by sending through the
  same backend send and is not changed by this feature.
- The Default version, ad-hoc (unsaved) requests and requests without aliases are unaffected.
- Constitution VI (Postman parity): aliases are already a recorded user decision (XEN-016); this
  feature extends that decision at the user's request (issue #25).
