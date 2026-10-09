# Data Model: Git Sync (Stage 7)

## Workspace file (`workspace.yaml` → `sync`, shareable)
```yaml
sync:
  repository: "git@github.com:acme/backend.git"   # or https URL
  branch: main
  subfolder: api/xenon                             # "" = repo root
```

## Local data
```
Xenon/
├── app.json                       ← gains gitAccounts: [GitAccount] (no secrets)
├── ssh/known_hosts                ← Xenon's own, pre-seeded
├── ssh/xenon_ed25519(.pub)        ← generated keys (0600)
├── repos/<localId>/               ← default clone location (own clone per workspace)
└── workspaces/<localId>/state.json ← gains sync: LocalSyncSettings
```

### GitAccount
| Field | Type | Notes |
|---|---|---|
| ID | ULID | |
| Kind | enum `github \| token \| ssh` | |
| DisplayName | string | e.g., "luis @ github.com" |
| Host | string | github.com, gitlab.acme.test, … |
| Username | string | host login (github/token) |
| CommitName / CommitEmail | string | author identity; editable |
| SSHMode | enum `agent \| key_file \| generated` | ssh only |
| KeyPath | string | key_file/generated |
| CredentialRef | string | keychain key `dev.brienze.xenon.git/<id>` (token or key passphrase) |

### LocalSyncSettings
`{ AccountID, ClonePath, AutoPull, AutoPush }`.

### SyncStatus (in memory, emitted)
`{ State: up_to_date | local_changes | ahead | behind | diverged | conflict | error | syncing |
not_linked, LocalChanges, Ahead, Behind, LastSyncAt, LastError: { code, message, next_step } }`.

### ChangeItem
`{ ItemID, ItemName, ItemKind (request|folder|collection|environment|workspace_settings),
Change (added|modified|deleted|renamed), Paths[] }` — paths mapped to Xenon items via the
XEN-001 id index (files read from the relevant tree for deleted items).

### ConflictItem
`{ ItemID, ItemName, ItemKind, Kind (both_modified|delete_modify|field), Fields?[] (for
structural files: path, mine, theirs), Mine, Theirs, Base (rendered DTOs), Resolution
(mine|theirs|both|keep|delete) }`.

### SyncSession (pull in progress with conflicts)
`{ ID, LocalHead, RemoteHead, Base, MergedEntries, Conflicts[], StartedAt }` — in memory; Cancel
discards; app quit discards (nothing was written).

## State machine (workspace sync)
```
not_linked --Link/Clone--> up_to_date
up_to_date --save--> local_changes --commit--> ahead --push--> up_to_date
* --fetch shows remote ahead--> behind --pull (FF)--> up_to_date
ahead + behind = diverged --pull--> (no conflicts) ahead | (conflicts) conflict --Finish--> ahead
conflict --Cancel--> diverged
any --error--> error (retry → previous state)
```
