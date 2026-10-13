# Feature Specification: Export (Stage 6)

**Feature Branch**: `xen-006`

**Created**: 2026-10-09

**Status**: Draft

**Input**: User description: "Xenon Stage 6 — Export at every level as Xenon YAML and Postman JSON, Xenon YAML import, passphrase-encrypted secrets for Xenon YAML, current-values option, Postman round-trip fidelity. Source: docs/prd/xenon-mvp-prd.md 'Import and export'; builds on XEN-001..005."

## Context

Stage 6 of 7. The PRD promises that everything can go out and come back in, at any level, as
Xenon YAML or Postman JSON, for free. Stage 3 imported Postman/Bruno/cURL and kept everything a
Postman export needs (passthrough); this stage writes it back out, adds Xenon YAML export and
import, and proves Postman → Xenon → Postman loses nothing.

## Clarifications

### Session 2026-10-09

- Q: "Include secrets" for Postman exports? → A: Never. Postman exports contain secret names only
  (type secret, empty value). "Include secrets" (passphrase-encrypted) exists for Xenon YAML
  exports only.
- Q: Initial or current values? → A: The export dialog asks each time ("Export current values as
  initial values", default off).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Export to Postman (Priority: P1)

Luis needs to hand a collection to a colleague who uses Postman, or keep a Postman copy. He
right-clicks a collection, folder, request or environment (or uses File → Export for the whole
workspace) and exports Postman JSON that Postman opens exactly as he sees it in Xenon.

**Why this priority**: Free Postman export is a core reason for Xenon (Postman paywalls it) and
the fidelity promise.

**Independent Test**: Import the stage-3 fixture corpus, export it back to Postman, compare with
the originals under the normalisation rules; import the export in Postman (or Newman) and send.

**Acceptance Scenarios**:

1. **Given** a collection, **When** Luis exports it as Postman, **Then** he gets a Postman
   collection v2.1 JSON file with the same folders, requests, order, auth, scripts, variables,
   descriptions and examples.
2. **Given** a folder, **Then** the export is a v2.1 collection containing only that folder (named
   after it), with the collection's auth, scripts and variables applied as Postman would inherit
   them (copied onto the exported folder level); **given** a request, **then** a v2.1 collection
   with that one request.
3. **Given** an environment, **Then** the export is a Postman environment JSON; secret variables
   appear with type `secret` and an empty value; plain variables have their initial values (or
   current values if he ticked the option).
4. **Given** the whole workspace, **Then** the export is a zip with one collection JSON per
   collection and one environment JSON per environment, plus a globals JSON.
5. **Given** Xenon-only settings (inherit-parent-scripts off, confirm-before-send, environment
   colour, Xenon-only auth fields, Xenon's `{{vault:…}}` mapping), **Then** they are dropped and
   the export summary lists each one with the item it applied to.
6. **Given** a collection that was imported from Postman, **When** exported back, **Then** the
   result equals the original under the normalisation rules (FR-014), including fields Xenon
   never showed (passthrough).

---

### User Story 2 - Export and share as Xenon YAML (Priority: P1)

Luis exports a workspace, collection, folder, request or environment as Xenon YAML to back it up,
send it, or move it to another workspace. A workspace export is a zip of its folder; collections
and folders can be a zip or a single bundled YAML file; a request or environment is its YAML file.

**Why this priority**: Moving items between workspaces (PRD: "use export and import") and backups
depend on it.

**Independent Test**: Export each level, import into an empty workspace, compare trees and files.

**Acceptance Scenarios**:

1. **Given** a workspace, **When** exported as Xenon YAML, **Then** the zip contains the workspace
   folder's files exactly (minus local-only/staging files; environment and variable files differ
   only when "Export current values as initial values" is on), and no secret value, current
   value, token or cookie.
2. **Given** a collection or folder, **Then** Luis chooses zip (folder structure) or a single
   bundled `.yaml` file containing the whole subtree.
3. **Given** a request or environment, **Then** the export is its YAML file.
4. **Given** "Include secrets", **Then** Xenon asks for a passphrase (twice) and the export carries
   the secret values of the exported items encrypted with it; without the option, secret values
   are absent and the secrets arrive as "missing".
5. **Given** "Export current values as initial values", **Then** plain variables' current values
   replace initial values in the exported files (the workspace itself is not changed).

---

### User Story 3 - Import Xenon YAML (Priority: P1)

Luis imports a Xenon YAML export (workspace zip or folder, collection zip or bundle, folder,
request, environment) into a workspace, with the same dialog, clash handling and summary as
stage 3. If the export has encrypted secrets, he enters the passphrase and they go into the target
workspace's vault.

**Why this priority**: Export is only useful if it comes back in; it is how items move between
workspaces.

**Independent Test**: Round-trip every export kind through import into another workspace,
including encrypted secrets with right and wrong passphrases.

**Acceptance Scenarios**:

1. **Given** a workspace zip, **When** imported as a new workspace (choose a folder), **Then** it
   is registered and identical to the source (apart from new local state).
2. **Given** a collection/folder/request/environment export, **When** imported into an existing
   workspace at a chosen place, **Then** it appears with new ids, clash handling as in XEN-003
   FR-018.
3. **Given** an export with encrypted secrets, **When** Luis enters the right passphrase,
   **Then** the secret values land in the target workspace's vault; a wrong passphrase is
   rejected (retry or "Import without secrets", which marks them missing).
4. **Given** a Xenon export from a newer format version, **Then** import is refused with a clear
   message (XEN-001 format §10).

---

### User Story 4 - Export dialog and summary (Priority: P2)

Every item's right-click menu has "Export…", and File → Export exports the workspace. The dialog
offers format (Xenon YAML / Postman), shape (zip / bundled file where applicable), "Include
secrets" (Xenon YAML only), "Export current values", then a save location; a summary lists what
was exported and every dropped or changed field.

**Why this priority**: Makes the above discoverable; exports work without the polish.

**Independent Test**: Walk the dialog for each item kind and format; check the summary.

**Acceptance Scenarios**:

1. **Given** Postman format is selected, **Then** "Include secrets" is not offered, with a short
   explanation.
2. **Given** an export finished, **Then** the summary shows counts and every dropped Xenon-only
   field or masked/omitted secret, and "Show in folder".
3. **Given** the target file exists, **Then** Luis confirms overwrite or picks another name.

---

### Edge Cases

- **Requests with binary example bodies** (base64 in Xenon): Postman v2.1 cannot hold binary
  example bodies; they are exported as empty bodies with a summary note.
- **Unsupported auth (stage 3)** exports back to its original Postman type and attributes; Bruno
  `unsupported` auth exports as `noauth` with a note.
- **Bruno-imported items** exported to Postman: Bruno-only fields (`bruno:` passthrough) are
  dropped with a note.
- **`srcState: repick` file references**: exported with their stored path (Postman keeps paths).
- **Disabled scripts** (`scripts.disabled`) export as Postman `event.disabled: true`.
- **OAuth tokens and cookies** are never exported (any format); OAuth configuration is.
- **Secrets in auth fields typed literally** (not via a secret variable) are exported as typed —
  they are not secret variables; the export summary warns when an auth field contains a literal
  value (non-`{{…}}`) so Luis can decide.
- **Empty collections/folders** export as empty `item: []`.
- **Large workspaces** (thousands of requests): export streams to the file with progress and can be
  cancelled; a cancelled export leaves no partial file.
- **Paths/names** in zips are portable (forward slashes, no absolute paths, Unicode NFC).

## Requirements *(mandatory)*

### Functional Requirements

**Postman export**

- **FR-001**: Users MUST be able to export a collection, folder or request as Postman collection
  v2.1 JSON, an environment as Postman environment JSON, the workspace globals as Postman globals
  JSON, and the workspace as a zip of these.
- **FR-002**: Postman export MUST invert the stage-3 mapping (XEN-003 postman-mapping.md):
  generate every mapped field from Xenon data, then add the item's Postman passthrough keys that
  Xenon does not own; Xenon-owned arrays (headers, query, form fields, variables, events, examples)
  are always regenerated so Xenon edits win.
- **FR-003**: Folder and request exports MUST carry the inherited collection/folder auth, scripts
  and collection variables they depend on, placed at the exported root, so the export behaves the
  same in Postman.
- **FR-004**: Postman exports MUST NOT contain secret values: secret rows are exported with
  `type: secret` and an empty value; OAuth tokens and cookies are never exported.
- **FR-005**: Xenon-only data with no Postman equivalent MUST be dropped and listed in the export
  summary: inherit-parent-scripts off, confirm-before-send, environment colour, Xenon `useToken`/
  `autoRefresh` OAuth fields, `bruno:` passthrough, binary example bodies.
- **FR-006**: Items created in Xenon (no passthrough) MUST export to valid Postman v2.1 JSON that
  Postman and Newman accept.

**Xenon YAML export**

- **FR-007**: Users MUST be able to export a workspace (zip of its folder excluding local-only and
  staging files), a collection or folder (zip of its directory, or one bundled YAML file), a
  request (its YAML file) and an environment (its YAML file).
- **FR-008**: Exports MUST NOT contain current values, tokens, cookies or secret values unless the
  user opts in per FR-009/FR-010.
- **FR-009**: "Include secrets" (Xenon YAML only) MUST ask for a passphrase (twice, minimum 8
  characters) and add the exported items' secret values encrypted with it; the encrypted part is
  self-contained (no keychain or vault key involved).
- **FR-010**: "Export current values as initial values" MUST be asked in the export dialog each
  time (default off) and, when on, write plain variables' current values as initial values in the
  export only.

**Xenon YAML import**

- **FR-011**: The Import dialog (XEN-003) MUST accept Xenon YAML exports: workspace zip or folder
  (imported as a new workspace at a chosen folder), collection zip or bundle, folder zip or
  bundle, request file, environment file — with XEN-003's planning, clash handling, atomicity and
  summary.
- **FR-012**: Imports with encrypted secrets MUST ask for the passphrase; correct → values go to
  the target workspace vault; wrong → retry or import without secrets (missing).
- **FR-013**: Imported items MUST get new ids; workspace-zip imports keep the workspace's own
  files as they are (ids included) since they become a new, separate workspace.

**Fidelity**

- **FR-014**: Postman → Xenon → Postman MUST reproduce the original collection/environment under
  these normalisation rules: object key order ignored; defaults written by Postman vs absent are
  equal (`disabled: false`, empty `description`); `url` compared by `raw` plus query/variable
  arrays; `info._postman_id` preserved; `_exporter_id`, `_postman_exported_at`,
  `_postman_exported_using` and secret values excluded from comparison.
- **FR-015**: Xenon YAML → import → Xenon YAML MUST be byte-identical for request, folder,
  collection and environment files (ids aside for non-workspace imports).

**Dialog and summary**

- **FR-016**: Every collection, folder, request and environment MUST have "Export…" in its
  right-click menu, and File → Export MUST export the workspace; the dialog offers format, shape,
  include secrets (Xenon YAML only), current values, then save location.
- **FR-017**: The export summary MUST list counts, every dropped/changed field with its item,
  literal-credential warnings, and offer "Show in folder".
- **FR-018**: Exports MUST be written atomically (temp file then rename), show progress over 200
  items, and be cancellable without leaving partial files.

### Key Entities

- **Export request**: item (kind + id), format (xenon_yaml | postman), shape (zip | bundle | file),
  include secrets (+ passphrase), export current values.
- **Bundle**: single YAML file holding a collection or folder subtree.
- **Encrypted secrets block**: passphrase-encrypted map of the exported items' secret values.
- **Export summary**: counts, dropped fields, warnings.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the stage-3 Postman fixture corpus round-trips (FR-014) with zero
  differences.
- **SC-002**: Exports of collections created in Xenon import into Postman/Newman without errors and
  send the same requests (Newman run against a test server) for 100% of a corpus covering all
  request features.
- **SC-003**: Xenon YAML exports re-import byte-identically (FR-015) for 100% of files.
- **SC-004**: 0 secret values in any Postman export and in any Xenon export without "Include
  secrets"; with it, values are unreadable without the passphrase.
- **SC-005**: Exporting a 1,000-request workspace takes under 5 seconds.

## Assumptions

- Postman collection v2.1 is the export target (v2.0 and v3 are not produced).
- Bundled YAML and encrypted-secrets blocks are Xenon formats (defined in the plan's contract).
- Moving items between workspaces is done by export + import (PRD); a direct "Move to workspace"
  action is post-MVP.
