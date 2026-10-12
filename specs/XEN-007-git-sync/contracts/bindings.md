# Contract: Go ↔ UI bindings — stage 7 additions

Prefixes: `GIT` sync · `ACC` accounts.

## 1. `GitAccountService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `List()` | — | `GitAccount[]` (no secrets) | |
| `StartGitHubSignIn()` | — | `{ user_code, verification_uri, expires_in, interval, session_id }` | opens browser to `verification_uri`; `ACC-01503` client id not configured |
| `PollGitHubSignIn(req)` | `{ session_id }` | `{ status: pending\|done\|expired\|denied, account? }` | UI polls per interval (Go enforces slow_down) |
| `AddToken(req)` | `{ provider: github\|gitlab\|bitbucket\|azure_devops\|other, host, username?, token, repository?, display_name?, commit_name, commit_email }` | `GitAccount` | verified with a ref advertisement of `repository` (`ACC-02401` invalid token, `ACC-18400` repository on another host) or, for GitHub without a repository, the user API (`ACC-17401`); other hosts without a repository `ACC-03400`; vault locked `ACC-16409` |
| `AddSSH(req)` | `{ mode: agent\|key_file\|generate, key_path?, passphrase?, host, commit_name, commit_email }` | `{ account: GitAccount, public_key }` | `ACC-10400` key file unreadable, `ACC-42400` wrong passphrase, `ACC-11400` no agent, `ACC-53400` agent without keys |
| `UploadPublicKey(req)` | `{ ssh_account_id, github_account_id, title }` | — | GitHub API |
| `Update(req)` / `Remove(req)` | `{ id, … }` | `GitAccount` / — | removal deletes keychain entries |
| `TestConnection(req)` | `{ account_id, repository }` | `{ default_branch, branches }` | unknown host key `ACC-05409` (+ event), changed `ACC-06409` |
| `ConfirmHostKey(req)` | `{ repository, fingerprint, accept }` | — | answers a `git:host-key-unknown` event; rescans and remembers only if the fingerprint still matches (`ACC-15400`) |

## 2. `SyncService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Status(req)` | `{ local_id }` | `SyncStatus` | |
| `Link(req)` | `{ local_id, repository, branch, subfolder, account_id, clone_path, mode?: adopt_repo\|replace_with_mine }` | `SyncStatus` | `GIT-01409` subfolder has a workspace and no mode; `GIT-02415` LFS in subfolder; `GIT-03404` branch missing |
| `CloneAsWorkspace(req)` | `{ repository, branch?, subfolder, account_id, clone_path, name? }` | `WorkspaceSummary` | |
| `Changes(req)` | `{ local_id }` | `ChangeItem[]` | |
| `Commit(req)` | `{ local_id, message, item_ids[] }` | `SyncStatus` | `GIT-00400` empty message |
| `Push(req)` | `{ local_id }` | `SyncStatus` | auto pull-and-retry once on non-FF |
| `Pull(req)` | `{ local_id }` | `{ status, session?: { id, conflicts: ConflictItem[] } }` | |
| `Resolve(req)` | `{ session_id, item_id, resolution, field_path? }` | `ConflictItem` | |
| `FinishPull(req)` / `CancelPull(req)` | `{ session_id }` | `SyncStatus` | finish refused while unresolved (`GIT-04409`) |
| `SaveSettings(req)` | `{ local_id, account_id, auto_pull, auto_push }` | — | |
| `Unlink(req)` | `{ local_id, target_folder }` | `WorkspaceSummary` | |

Any call while another sync operation runs on the same workspace returns `GIT-05409` (busy).

Errors carry `next_step` in `error_details`: `GIT-10503` offline, `GIT-11401` auth expired,
`GIT-12403` org approval / SSO required, `GIT-13409` push rejected (other), `GIT-14403` protected
branch / signature required.

### Implementation notes (T019)
- `SyncStatus` = `{ state, local_changes, ahead, behind, last_sync_at?, last_error?: { code, message,
  next_step }, notice?, settings: { repository, branch, subfolder, account_id, clone_path, auto_pull,
  auto_push } }`; `notice` reports a `workspace.yaml` `sync` section that differs from the linked one.
- `Status` fetches; a remote failure (or a missing account, `GIT-07428` "No account - choose one")
  comes back as `state: error` with `last_error`, not as a failed call; while another operation runs
  it answers `state: syncing` instead of `GIT-05409`.
- `Link`: `clone_path` optional (default `<app data>/repos/<local_id>`), validation `GIT-63400`
  (incl. a subfolder escaping the repository), clone path not empty `GIT-65409`, overlapping a
  workspace `GIT-66409`.
- `Commit`: validation `GIT-00400`, none of the items changed `GIT-08400`; operations on a workspace
  without a repository `GIT-09409`; Commit/Push/Pull without an account `GIT-07428`.
- `Pull` with conflicts (US1): `{ status: conflict, session: { id, conflicts: [{ item_id, item_name,
  item_kind, kind, resolution, path }] } }`, nothing written; `Push` behind the remote pulls,
  replays and retries once (`GIT-21409` only if the retry is refused again), a conflicting replay
  answers `state: conflict`.
- `Changes` items: `{ item_id, item_name, item_kind, change, paths }`; workspace-level files
  (`.gitignore`) belong to the workspace item.

### Implementation notes (T022)
- `Pull` (and `Push` when its replay conflicts) keeps a session per workspace, in memory: `Status`
  answers `state: conflict` while it is open; a new pull replaces it; a pull without conflicts drops it.
- `Session { local_id }` → `{ id, local_id, conflicts }` or null (extra method, for reopening the view).
- `ConflictItem` = `{ item_id, item_name, item_kind, kind, resolution, path, uncommitted, base?, mine?,
  theirs?, fields: [{ path, mine?, theirs?, resolution }] }`; a version is `{ text, request?, environment? }`
  (request = the `Request` DTO; environment = `{ id, name, description, color, confirm_before_send,
  variables: [{ key, value, description, secret, enabled }] }`, secret values never included); absent
  side = deleted there; a field value null = removed on that side.
- `Resolve`: `resolution` mine|theirs|both|keep|delete (validation `GIT-29400`); `both` only for requests
  and environments, keep/delete only for delete-vs-modify, `field_path` (mine|theirs) only for field
  conflicts → otherwise `GIT-19400`; unknown item `GIT-35404`; closed or replaced session `GIT-16404`
  (`GIT-38404` when it is replaced while resolving); a file whose fields are all resolved answers
  `resolution: merged`.
- `FinishPull`: unresolved `GIT-04409`; branch moved since the pull `GIT-28409` (session dropped, pull
  again); returns the new `SyncStatus` (usually `ahead`, `up_to_date` when the result is the remote).
  `CancelPull` returns the status as before the pull. Both are sync operations (`GIT-05409` while one runs).

## 3. Events
`git:status-changed { local_id, status }`, `git:progress { local_id, phase, done, total }`,
`git:host-key-unknown { repository, host, fingerprint, key_type }`.
