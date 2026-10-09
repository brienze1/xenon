# Implementation Plan: Git Sync (Stage 7)

**Branch**: `xen-007` (stacked on `xen-006`) | **Date**: 2026-10-09 | **Spec**: [spec.md](spec.md)

## Summary

Per-workspace Git sync on go-git v5 with Xenon's own sparse clone, a Xenon-specific sync engine
(fetch + path-level 3-way + semantic merge of structural files + linear replay, since go-git has
no merge/rebase), per-item conflict resolution, Git accounts (GitHub device flow, tokens with host
conventions, SSH agent/key/generated) with credentials in the OS keychain and Xenon's own
known_hosts, footer status, Sync panel, auto pull/push and clone-as-workspace.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 5 + React 19

**Primary Dependencies** (new): `github.com/go-git/go-git/v5` v5.19.3 (+ its transport/ssh);
`golang.org/x/crypto/ssh` (key generation); tests: `github.com/gliderlabs/ssh` (in-process SSH
server)

**Storage**: Xenon clones under app data (or user-chosen paths); accounts in `app.json` +
keychain; `ssh/known_hosts`, generated keys

**Testing**: godog BDD with local bare repos over file://, in-process smart-HTTP and SSH servers,
two Xenon clones for concurrency, mock GitHub device-flow/API; property-based test for the merge
engine (random concurrent edits — SC-002); Vitest + RTL

**Target Platform / Project Type**: desktop-app, macOS/Windows/Linux

**Performance Goals**: SC-001 (commit+push / pull < 5 s typical)

**Constraints**: never touch other checkouts; never commit secrets/local state; one operation per
repo at a time

**Scale/Scope**: repos up to large monorepos (sparse checkout limits worktree size)

## Constitution Check

- [x] **I. Local-first** — Git is optional and talks only to the user's hosts; no Xenon server
  (device flow needs only a public client id). ✅
- [x] **II. Secrets** — tokens/passphrases in keychain; never in clone config or remote URLs;
  only the workspace folder is committed, which never contains secrets (stages 1–2); SC-003
  verified on history. ✅
- [x] **III. Go owns logic** — all Git, merge and auth in Go. ✅
- [x] **IV. Architecture skills** — Git operations are integration webservices/persistence of an
  external system; browser/keychain via existing infra adapters. ✅
- [x] **V. BDD-first** — features per story with real Git over local servers. ✅
- [x] **VI. Postman parity** — n/a (Postman's Git is paid/cloud); scope = PRD. ✅
- [x] **VII. Cross-platform** — SSH agent per OS (Pageant/OpenSSH pipe), LF writes, CI on 3 OSes. ✅

**Post-design re-check**: all gates pass.

## Project Structure

### Documentation
```text
specs/XEN-007-git-sync/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/bindings.md
├── checklists/
└── tasks.md
```

### Source code (additions)
```text
internal/domain/entity/  git_account.go, sync_settings.go, local_sync_settings.go, sync_status.go,
                         change_item.go, conflict_item.go, sync_session.go
internal/domain/enums/   git_account_kind, ssh_mode, sync_state, change_type, conflict_kind,
                         conflict_resolution (+ value pkgs)
internal/application/
├── usecase/  pure: classify_divergence, plan_three_way (path classification), merge_structural_file
│             (R3), squash_replay_message, map_paths_to_items; ports: clone_repo, fetch_branch,
│             read_tree_versions, write_sync_commit, push_branch, find/save_git_account,
│             find/save_git_credential, verify_token, github_device_flow, generate_ssh_key,
│             verify_host_key
├── adapter/  sync_service, git_account_service
└── service/  sync.go (link, clone, status, commit, push, pull, sessions, auto pull/push),
              git_accounts.go
internal/integration/
├── webservice/git/      repo.go (go-git open/clone/sparse/fetch/push with per-repo lock),
│                        trees.go (read versions, build trees, commit), auth.go (BasicAuth per host,
│                        SSH agent/key), hostkeys.go (known_hosts + pre-seeded keys)
├── webservice/github/   device_flow.go, api.go (user, emails, ssh keys)
├── persistence/         git_accounts.go (app.json section), state.json sync settings
└── entrypoint/controller/ sync_*.go, git_account_*.go
internal/infra/server/binding/ sync.go, git_account.go; event/events.go adds git events
test/integration/
├── features/git/  link_and_clone, status_and_changes, commit_push_pull, fast_forward,
│                  divergence_replay, conflicts_requests, conflicts_structural,
│                  conflicts_environments, uncommitted_preserved, accounts_github_device,
│                  accounts_token_hosts, accounts_ssh, host_keys, errors_next_steps,
│                  auto_pull_push, unlink, never_touch_outside, secrets_never_committed
├── mock/          git_servers.go (bare repos + smart-HTTP + SSH), github_api.go
└── property/      merge_engine_property_test.go (100 random concurrent edits)
frontend/src/features/sync/  store.ts, FooterSyncStatus, SyncPanel (changes, message, actions,
                             errors with next step), ConflictView (side-by-side item renderers,
                             field conflicts), LinkRepoDialog, CloneWorkspaceDialog
frontend/src/features/settings/ GitAccounts (GitHub device code UI, token form, SSH options,
                             public key copy/upload), HostKeyPrompt
```

**Structure Decision**: Git remotes and GitHub are external systems → integration webservices;
the merge engine's rules are pure application use cases operating on domain entities, so they are
unit/property-testable without Git.

## Key design decisions
1. Own clone, sparse, full history, single branch; branch change = re-clone.
2. Linear replay (one squashed sync commit on top of remote) instead of merge commits.
3. Structural YAML merged semantically; requests/environments resolved per item.
4. Uncommitted local work is never lost: snapshot → update → re-apply; overlap → conflict.
5. Per-repo mutex; operations queued; UI disabled while running.
6. Credentials only via keychain-backed account references; remote URLs stored without
   credentials.

## Complexity Tracking

No constitution violations to justify.
