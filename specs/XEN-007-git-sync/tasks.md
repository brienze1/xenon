---
description: "Task list for XEN-007 — Git Sync (Stage 7)"
---

# Tasks: Git Sync (Stage 7)

**Input**: `specs/XEN-007-git-sync/` · **Prerequisite**: XEN-001..006 implemented; GitHub OAuth App
client id provided by the owner before T029 (real GitHub validation; BDD uses a mock — spec Assumptions).

**Tests**: MANDATORY. Feature files tagged `@all @xen_007 @git`, written first and proven red.

US1 link & daily sync · US2 conflicts · US3 accounts · US4 clone as workspace · US5 auto pull/push

---

## Phase 1: Setup

- [ ] T001 Add `github.com/go-git/go-git/v5@v5.19.3` and `golang.org/x/crypto` (ssh) to `go.mod`; test-only `github.com/gliderlabs/ssh`; properties `XENON_GITHUB_CLIENT_ID` (build-time default), `XENON_GITHUB_API_URL`, `XENON_GITHUB_LOGIN_URL`
- [ ] T002 [P] Test infrastructure `test/integration/mock/git_servers.go`: create bare repos, serve over file://, in-process smart-HTTP (BasicAuth checks per host convention) and in-process SSH (key auth, known host key), helpers to commit to the remote from a second client; `mock/github_api.go` (device code/token polling incl. `slow_down`, `expired_token`, `access_denied`; `/user`, `/user/emails`, `/user/keys`)
- [ ] T003 [P] Generic steps `given_git_steps_test.go` / `then_git_steps_test.go`: `the remote "<name>" has file "<path>" with` (docstring), `a teammate changes "<path>" on remote "<name>" to` (docstring), `the remote "<name>" history should not contain "<text>"`, `the remote "<name>" branch "<b>" should have <N> commits`, `the file "<path>" on the remote should equal` (docstring)

---

## Phase 2: Foundational

- [ ] T004 [P] Enums + entities from data-model.md in `internal/domain/`
- [ ] T005 [P] Pure use cases with table tests: `classify_divergence.go`, `plan_three_way.go`, `merge_structural_file.go` (research R3), `squash_replay_message.go`, `map_paths_to_items.go`
- [ ] T006 Property test `test/integration/property/merge_engine_property_test.go`: 100 random concurrent edit pairs on generated workspaces → no change lost, every overlap reported as conflict (SC-002)
- [ ] T007 `internal/integration/webservice/git/repo.go` (open/clone single-branch full history + sparse checkout, fetch, push with ForceWithLease, per-repo mutex), `trees.go` (read base/mine/theirs versions, build tree with replaced subfolder subtree, commit, update ref, checkout), `auth.go`, `hostkeys.go` (own known_hosts, pre-seeded github.com/gitlab.com/bitbucket.org keys, TOFU via event + `ConfirmHostKey`) with unit tests against T002 servers

**Checkpoint**: engine and Git plumbing tested in isolation.

---

## Phase 3: User Story 3 — Accounts (P1)

- [ ] T008 [P] [US3] BDD `test/integration/features/git/accounts_token_hosts.feature`: token accounts for GitHub/GitLab/Bitbucket/Azure conventions verified against the smart-HTTP mock; invalid token `ACC-02401`; token stored in keychain mock only (not in app.json, clone config or remote URL); removing an account deletes its keychain entries and linked workspaces report "No account — choose one" (FR-016)
- [ ] T009 [P] [US3] BDD `test/integration/features/git/accounts_ssh.feature` and `host_keys.feature`: agent (mock agent socket), key file with passphrase, generated ed25519 (0600, public key returned, upload to GitHub mock); pre-trusted hosts; unknown host → event + confirm; changed key blocked
- [ ] T010 [P] [US3] BDD `test/integration/features/git/accounts_github_device.feature`: start → code + URI (browser opener mock), polling pending → done → account with username/email; `slow_down`, `expired_token`, `access_denied`; missing client id `ACC-01503`
- [ ] T011 [US3] Run T008–T010 → **red**; implement `webservice/github/{device_flow,api}.go`, `persistence/git_accounts.go`, keychain refs (XEN-002 keychain adapter, service `dev.brienze.xenon.git`), `service/git_accounts.go`, controllers/bindings (bindings §1) → **green**
- [ ] T012 [US3] UI `features/settings/GitAccounts` (list; GitHub device code screen with copy + open browser + status; token form with host presets; SSH options with public key copy/upload), `HostKeyPrompt`; RTL tests

---

## Phase 4: User Story 1 — Link and daily sync (P1) 🎯 MVP

- [ ] T013 [P] [US1] BDD `test/integration/features/git/link_and_clone.feature`: link with empty/missing subfolder (files placed, first commit offered), subfolder with existing workspace (adopt / replace), LFS refused `GIT-02415`, branch missing `GIT-03404`; registry re-pointed, old folder untouched; only subfolder checked out
- [ ] T014 [P] [US1] BDD `test/integration/features/git/status_and_changes.feature` and `commit_push_pull.feature`: status states and counts; changes mapped to item names (added/modified/deleted/renamed); commit only checked items under subfolder with account identity; push; teammate change → behind → pull (FF) → items updated and clean tabs reload event
- [ ] T015 [P] [US1] BDD `test/integration/features/git/divergence_replay.feature` and `uncommitted_preserved.feature`: diverged non-overlapping → one squashed commit on top of remote with original messages in body; push after; uncommitted local edits preserved through FF and replay; overlap with uncommitted → conflict
- [ ] T016 [P] [US1] BDD `test/integration/features/git/never_touch_outside.feature` and `secrets_never_committed.feature`: a separate checkout of the same remote is byte-identical before/after a session (SC-004); after secrets/current values/tabs/history/cookies exist, the remote history contains none of their values (SC-003)
- [ ] T017 [P] [US1] BDD `test/integration/features/git/errors_next_steps.feature`: offline `GIT-10503`, auth expired `GIT-11401`, org approval/SSO `GIT-12403` (mock 403 bodies), protected branch/signature `GIT-14403`, push non-FF → auto pull + retry once; a second sync operation on the same workspace while one runs is queued/refused (`GIT-05409`) and never interleaves (FR-020)
- [ ] T018 [US1] Run T013–T017 → **red**
- [ ] T019 [US1] Implement `service/sync.go` (link, status with 5-min/focus refresh trigger from UI, changes, commit, push, pull FF/replay without conflicts, uncommitted snapshot/re-apply, error mapping with next steps; files written by pulls registered as own writes so the XEN-001 watcher emits a single `workspace:tree-changed` instead of external-change banners; remote changes to `workspace.yaml` `sync` shown as a notice, not applied), controllers/bindings (bindings §2), events → **green**
- [ ] T020 [US1] UI `features/sync/`: `FooterSyncStatus`, `SyncPanel` (changes with checkboxes, message, Commit / Commit and Push / Push / Pull, progress, errors with next step), `LinkRepoDialog` (URL, branch, subfolder, account, clone location, adopt/replace choice); workspace settings Sync tab; RTL tests

---

## Phase 5: User Story 2 — Conflicts (P1)

- [ ] T021 [P] [US2] BDD `test/integration/features/git/conflicts_requests.feature`, `conflicts_structural.feature`, `conflicts_environments.feature`: both-modified request (mine/theirs/both → "<name> (theirs)" new id), delete-vs-modify (keep/delete), structural auto-merge of order lists and fields with field conflicts, environment row merge and item conflict; Finish commits on top of remote; Cancel restores the exact pre-pull state; unresolved Finish `GIT-04409`
- [ ] T022 [US2] Run → **red**; implement sync sessions (`Resolve`, `FinishPull`, `CancelPull`) on top of T005/T019 → **green**
- [ ] T023 [US2] UI `features/sync/ConflictView` (list, side-by-side read-only renderers reusing request/environment components, field conflicts for structural files, resolve in any order, Finish/Cancel); RTL tests

---

## Phase 6: User Story 4 — Clone as workspace (P2)

- [ ] T024 [P] [US4] BDD `test/integration/features/git/link_and_clone.feature` additions: `CloneAsWorkspace` with existing workspace subfolder (opened, secrets missing) and empty subfolder (new workspace created); default branch from remote
- [ ] T025 [US4] Run → **red**; implement `CloneAsWorkspace` → **green**; UI `CloneWorkspaceDialog` in the workspace switcher ("New workspace from Git repository"); RTL test

---

## Phase 7: User Story 5 — Auto pull/push, unlink (P3)

- [ ] T026 [P] [US5] BDD `test/integration/features/git/auto_pull_push.feature` and `unlink.feature`: auto-pull on activation (conflict → status only, nothing changed), auto-push after 3 s debounce with "Update <names>" (time mock), failures surfaced; unlink copies files to target folder, re-points workspace, deletes clone after confirmation
- [ ] T027 [US5] Run → **red**; implement auto triggers (activation hook, save debounce) and `Unlink` → **green**; UI toggles in Sync tab

---

## Phase 8: Polish

- [ ] T028 [P] Performance BDD: commit+push and conflict-free pull of 10 changed requests < 5 s against the local smart-HTTP server (SC-001)
- [ ] T029 Run a real end-to-end session against GitHub (test repo) with device-flow sign-in, token and SSH on macOS, Windows and Linux (SC-005/006); record results in `specs/XEN-007-git-sync/quickstart.md` notes
- [ ] T030 `/refactor-diff` until clean; `make validate`; CI on 3 OSes; walk [quickstart.md](quickstart.md)

---

## Dependencies & Execution Order
Setup → Foundational → US3 (accounts needed to authenticate) → US1 → US2 → US4 → US5 → Polish.
Only T029 (manual validation against real GitHub) needs the owner's GitHub client id; BDD uses the
mock.

## Parallel opportunities
T002–T003; T004–T006; all [P] BDD files per story; UI tasks alongside backend once bindings are
fixed.

## Implementation Strategy
MVP = US3 + US1 + US2 (sync works safely with conflicts) → US4 → US5 → Polish.
