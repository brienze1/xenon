# Contract: Go ↔ UI bindings — stage 7 additions

Prefixes: `GIT` sync · `ACC` accounts.

## 1. `GitAccountService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `List()` | — | `GitAccount[]` (no secrets) | |
| `StartGitHubSignIn()` | — | `{ user_code, verification_uri, expires_in, interval, session_id }` | opens browser to `verification_uri`; `ACC-01503` client id not configured |
| `PollGitHubSignIn(req)` | `{ session_id }` | `{ status: pending\|done\|expired\|denied, account? }` | UI polls per interval (Go enforces slow_down) |
| `AddToken(req)` | `{ host, username?, token, commit_name, commit_email }` | `GitAccount` | verified; `ACC-02401` invalid token |
| `AddSSH(req)` | `{ mode: agent\|key_file\|generate, key_path?, passphrase?, host, commit_name, commit_email }` | `GitAccount` + `{ public_key? }` | |
| `UploadPublicKey(req)` | `{ ssh_account_id, github_account_id, title }` | — | GitHub API |
| `Update(req)` / `Remove(req)` | `{ id, … }` | `GitAccount` / — | removal deletes keychain entries |
| `ConfirmHostKey(req)` | `{ host, fingerprint, accept }` | — | answers a `git:host-key-unknown` event |

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

## 3. Events
`git:status-changed { local_id, status }`, `git:progress { local_id, phase, done, total }`,
`git:host-key-unknown { host, fingerprint, key_type }`.
