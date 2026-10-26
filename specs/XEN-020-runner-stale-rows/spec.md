# Feature Specification: Runner Follows the Current Tree

**Feature Branch**: `xen-020`

**Created**: 2026-10-26

**Status**: Draft

**Input**: User description: "The collection/folder Runner must always show and run the current contents of its target. Today, once a folder's Runner has been opened in a Xenon session, its row list (run order + selection) stays frozen: requests or subfolders added to the folder afterwards (from the sidebar, the MCP tree_edit tool, a git pull or an external file change) do not appear in the Runner, and requests deleted from it linger, until Xenon is restarted. A run started from the Runner then silently skips the new requests. Users need the Runner to reconcile the remembered row list with the current tree whenever the tree changes or the Runner is shown again: new requests appear in tree order (after their preceding sidebar sibling), selected by default; removed requests disappear; the user's existing order, unchecked rows and alias choices are kept; nothing the user edited in the Runner is lost by the refresh. Headless runs started by AI clients (MCP run start without explicit rows) must likewise include every current request of the target. Reported for Xenon 0.1.9: a folder run once with 64 rows still showed 64 after 29 requests were added (expected 93). Source: https://github.com/brienze1/xenon/issues/32"

## Context

XEN-017 added a **Runner** section to every collection and folder tab. Its **Setup** view lists
the target's requests as rows (run order, a checkbox per row, an alias choice per row), and that
row list is remembered per collection/folder on this machine.

The bug (issue #32, Xenon 0.1.9): the Runner's row list is read **once per app session**, the
first time a target's Runner is shown. Later changes to the target's contents in the same session
are never reflected:

- Requests and subfolders added to the target do not appear as rows.
- Requests deleted from (or moved out of) the target stay as rows.
- A run started from the Runner sends exactly the rows shown, so it **silently skips** the new
  requests.

Restarting Xenon "fixes" it, because the first read of the new session reconciles the
remembered rows with the tree on disk (that reconciliation already exists and is correct: new
requests join selected, right after the nearest preceding request in sidebar order; removed requests drop out; order,
unchecked rows and alias choices are kept). The sidebar itself is up to date the whole time.

### Settled from the code rather than asked

| Question | Answer and the fact that settles it |
| --- | --- |
| Where does the staleness come from? | The Runner view keeps each target's rows from its first read and never reads them again in that session; a tree change refreshes the sidebar only. The remembered-rows store on disk is reconciled correctly whenever it is read. |
| Do headless (AI client) runs without explicit rows skip new requests? | No, not by themselves: a headless start reads the target's tree from disk and reconciles the remembered rows on every call. Reconciliation happens on every read, so even a frozen row list written back as the remembered rows is completed with the new requests at the next headless start. The report's headless observation (22 planned rows for 25 requests) is not reproducible from the code; the run counted was most likely started from the Runner, which sends its frozen rows (Story 2). This spec locks the headless behaviour with an acceptance scenario so it cannot regress. |
| Are new requests selected? | Yes — existing reconciliation selects every request the remembered rows do not know. |
| Where does a new request go in the run order? | Right after the nearest request preceding it in sidebar order (depth first, subfolders included) that is already a row; first when none of its preceding siblings is a row (existing reconciliation). |
| Do explicit rows given by an AI client get topped up with new requests? | No. Explicit rows are an explicit choice ("run exactly these"); only runs without explicit rows use the remembered, reconciled rows. Unchanged by this feature. |

## Clarifications

### Session 2026-10-26

No question needed the user; each was settled from the issue or the code:

- Q: Refresh an open Runner live on tree changes, or only when it is shown again? → A: Live, on every (coalesced) tree change of the active workspace (FR-001, FR-004). The issue proposes either; only the live refresh keeps a Runner that stays open from starting a run with stale rows, and it also covers "shown again".
- Q: Tell the user that rows were added or removed? → A: No notice; the rows simply match the tree, as on a fresh start of Xenon (XEN-017 shows no notice there either).
- Q: Re-read every Runner target known this session, or only Runners on screen? → A: Every target already read this session (FR-001); a target never read needs nothing (its first read is current). Bounded by the collection/folder tabs opened in one session.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The open Runner picks up tree changes (Priority: P1)

A user has a folder's Runner open (or opened earlier in this session). They, an AI client, a
git pull or an external editor add requests and subfolders to that folder, or delete requests
from it. Without restarting Xenon, the Runner's Setup rows match the folder's current contents:
new requests appear, selected, in tree order; removed ones disappear; everything the user set
in the Runner (order, unchecked rows, alias choices, run settings) is kept.

**Why this priority**: This is the reported bug; without it a run silently skips requests.

**Independent Test**: Open a folder's Runner (64 rows), add 6 requests to an existing subfolder
and a new subfolder with 23 requests, then look at the Runner again: 93 rows, the 29 new ones
checked and in tree order.

**Acceptance Scenarios**:

1. **Given** a folder's Runner was shown with 64 rows, **When** 6 requests are added to one of its subfolders and a new subfolder with 23 requests is added, **Then** the Runner shows 93 rows without a restart, the 29 new rows checked, each new row placed right after the nearest request preceding it in sidebar order.
2. **Given** a folder's Runner was shown with a row unchecked, two rows reordered and an alias chosen on a row, **When** a request is added to the folder, **Then** the unchecked row stays unchecked, the custom order stays, the alias choice stays, and the new request appears checked.
3. **Given** a folder's Runner was shown, **When** a request in that folder is deleted or moved to another folder, **Then** its row disappears from the Runner.
4. **Given** a folder's Runner tab is open, **When** the folder's contents change while the user is looking at Setup, **Then** the rows update in place without the user closing or reopening anything.
5. **Given** the user unchecked a row less than a moment before a tree change arrives, **When** the Runner refreshes, **Then** that row is still unchecked (the user's last edit is not lost).

---

### User Story 2 - Runs started from the Runner include current requests (Priority: P1)

A run started from the Runner sends the target's current selected requests, including requests
added since the Runner was first shown.

**Why this priority**: The visible symptom is rows; the harmful one is a run that skips requests
while reporting success.

**Independent Test**: After the tree change of Story 1, click Run; the run plans 93 requests.

**Acceptance Scenarios**:

1. **Given** the situation of Story 1 scenario 1, **When** the user starts the run from the Runner, **Then** the run sends all 93 checked requests in the Runner's order.
2. **Given** a request was deleted from the folder after the Runner was first shown, **When** the user starts a run, **Then** the deleted request is not part of the run.

---

### User Story 3 - Headless runs include current requests (Priority: P2)

An AI client starts a run on a collection or folder without giving explicit rows. The run uses
the remembered rows reconciled with the target's current contents, so requests added since the
remembered rows were last saved are included, checked.

**Why this priority**: Already correct in the engine; it needs a guard so the bug report's second
symptom can never happen.

**Independent Test**: Remember rows for a folder, add requests to it, start a headless run
without rows: the planned run includes the new requests.

**Acceptance Scenarios**:

1. **Given** a folder whose remembered rows list 22 requests, **When** 3 requests are added to the folder and an AI client starts a run on it without explicit rows, **Then** the run plans 25 requests.
2. **Given** remembered rows with one request unchecked, **When** a request is added and a headless run starts without explicit rows, **Then** the unchecked request is not run and the new one is.

---

### Edge Cases

- **Several changes in quick succession** (an AI client creating 29 requests one by one): the
  Runner ends showing the final contents; intermediate refreshes may be skipped but the last
  change is always reflected.
- **A pending Runner edit and a tree change at the same time**: the user's edit is remembered
  first, then the rows are refreshed, so the edit survives (Story 1, scenario 5).
- **A run is active on the target**: the Results view of the active run is not altered; the
  Setup rows refresh for the next run. The active run keeps the rows it started with.
- **The target itself is deleted, or the re-read fails for another reason**: the Runner shows
  its existing load-error banner with **Try Again** in place of the Setup rows (FR-010); it does
  not keep showing rows that can no longer be run, and nothing crashes.
- **A request renamed or its method changed**: its row shows the new name/method; position,
  checkbox and alias are kept.
- **An alias chosen on a row is deleted from the request**: the row falls back to Default
  (existing reconciliation).
- **Ignored items** (files the sidebar ignores) never become rows (existing behaviour).
- **Another workspace's tree changes**: Runners of the active workspace are not refreshed.
- **Every request removed from the target**: the Setup list is empty; starting a run is refused
  with the existing "Check at least one request in the run order." message (`RUN-03400`).
- **A run started in the instant between a tree change and the refresh landing**: it sends the
  rows shown at that moment (FR-006 applies to the rows shown); a request deleted meanwhile shows
  in that run as a failed "missing request" row, as it does today. The next run uses the
  refreshed rows.
- **The start confirmation is open** ("Confirm before send" environment) when the tree changes:
  confirming runs the configuration the user was asked to confirm; the refresh updates the Setup
  rows for the next run only.
- **Runner not shown yet this session**: nothing to refresh; the first showing reads the current
  contents (existing behaviour).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When the contents of the active workspace's tree change, every collection/folder Runner already read in this session MUST re-read its rows from the reconciled remembered rows, without a restart and without the user reopening it.
- **FR-002**: The refreshed rows MUST follow the existing reconciliation rules: requests not in the remembered rows join checked, right after the nearest request preceding them in sidebar order (depth first under the target, subfolders included) that is already a row (first when none is); requests no longer in the target drop out; the remembered order, unchecked rows and alias choices of the remaining rows are kept; an alias that no longer exists falls back to Default.
- **FR-003**: Any Runner edit not yet remembered when a refresh starts (row order, checkboxes, alias choices, run settings, data file) MUST be remembered before the rows are re-read, and an edit made while a re-read is in flight MUST NOT be overwritten by that re-read (the rows are read again after the edit is remembered), so a refresh never discards an edit.
- **FR-004**: Refreshes caused by a burst of tree changes MUST coalesce: at most one re-read per Runner target in flight, plus one trailing re-read when further changes arrived meanwhile, so the final state always reflects the last change.
- **FR-005**: A refresh MUST NOT change which view (Setup, Results, Runs) is shown, the run shown in Results, or any active run.
- **FR-006**: A run started from the Runner MUST send the rows shown after the latest refresh (Story 2).
- **FR-007**: A run started by an AI client without explicit rows MUST use the remembered rows reconciled with the target's current contents at the moment of the start (Story 3). Runs with explicit rows keep running exactly those rows.
- **FR-008**: Tree changes of a workspace other than the active one MUST NOT refresh any Runner, and a re-read that completes after the active workspace changed MUST be discarded.
- **FR-009**: The fix MUST NOT write the run order or selection to any git-tracked workspace file (remembered rows stay in local UI state, as in XEN-017).
- **FR-010**: If a refresh's re-read fails (for example the target was deleted), the Runner MUST stop showing the old rows and show its existing load-error banner with **Try Again**, exactly as a failed first read does.

### Key Entities

- **Runner target**: a collection or folder whose tab shows a Runner section.
- **Remembered rows**: per target, on this machine — the run order, each row's checkbox and alias choice.
- **Runner rows (shown)**: the remembered rows reconciled with the target's current requests; what Setup displays and what a Runner-started run sends.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the reported scenario (64 rows, 29 requests added) the Runner shows 93 rows with no restart, and a run started from it sends 93 requests.
- **SC-002**: After a tree change, an open Runner shows the new contents within 1 second of the sidebar showing them.
- **SC-003**: 0 user Runner edits (order, checkbox, alias, settings) are lost because of a refresh across all acceptance scenarios.
- **SC-004**: A headless run without explicit rows started after requests were added includes 100% of the target's checked current requests.

## Assumptions

- "Tree change" means every change that updates the sidebar's tree (sidebar edits made in Xenon, AI client tree edits, git pulls, imports, external file edits); no new change source is added.
- The existing reconciliation rules (XEN-017) are the desired rules; this feature only makes the Runner apply them again within a session.
- Rows of a Runner target that was never shown in this session need no refresh; its first showing already reads current contents.
- Reading the rows of the targets already shown in a session after each (coalesced) tree change is cheap enough at the documented scale (a folder of ~100 requests).
