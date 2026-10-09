# Data Model: Environments, Variables and Secrets (Stage 2)

Extends [XEN-001 data-model.md](../XEN-001-shell-and-storage/data-model.md). File shapes:
[contracts/format-additions.md](contracts/format-additions.md).

## Workspace data (shared via git)

### Environment (new entity, `internal/domain/entity/environment.go`)
| Field | Type | Rules |
|---|---|---|
| ID | ULID | immutable |
| Name | string | 1–200 chars, unique per workspace (case-insensitive) |
| Description | string | |
| Color | enum `none\|green\|cyan\|magenta\|yellow\|red` | default per §2 of the format additions |
| ConfirmBeforeSend | bool | |
| Variables | []VariableRow | |

### VariableRow (replaces XEN-001 KeyValue for variable tables)
| Field | Type | Rules |
|---|---|---|
| Key | string | exact (not trimmed); no `{`, `}`, line breaks; case-sensitive; editor warns on leading/trailing spaces |
| InitialValue | string | plain rows only; empty for secret rows |
| Secret | bool | |
| Enabled | bool | default true |
| Description | string | |

Owners: Environment, Collection (not Folder), Workspace (globals).

### Scope reference (`ScopeRef`)
Identifies a variable table: `env:<environmentId>`, `collection:<collectionId>`, `globals`,
`local` (stage 4, in memory only). Used as the key for current values and vault entries.

## Local data (per machine, app data `workspaces/<localId>/`)

```
workspaces/<localId>/
├── state.json          (XEN-001)
├── history.jsonl       (XEN-001)
├── variables.json      ← active environment + current values of plain rows
├── vault.age           ← encrypted secret values
└── vault-key.age       ← only in passphrase mode: the vault identity wrapped with the passphrase
```

### LocalVariables (`variables.json`)
```json
{
  "version": 1,
  "activeEnvironmentId": "01JA3B…",
  "current": {
    "env:01JA3B…": { "base_url": "http://localhost:8080" },
    "collection:01JA2Z…": { "page_size": "50" },
    "globals": { "user": "luis" }
  }
}
```
- A key present in `current` overrides the row's initial value; absent → initial value.
- Reset deletes the key; Persist writes it to the file and deletes the key.
- Renaming a row in Xenon moves its current value (and vault entry) to the new key in the same
  operation. Rows renamed outside Xenon lose their local override (documented behaviour).
- Owner-only permissions (XEN-001 local data rules).

### Vault (`vault.age`)
- Content (before encryption): `{ "version": 1, "secrets": { "<ScopeRef>": { "<key>": "<value>" } } }`.
- Encrypted to the workspace's **vault identity** (an X25519 key pair). Written atomically.
- The vault identity is stored either:
  - **keychain mode**: in the OS keychain, service `dev.brienze.xenon.vault`, account
    `<localId>`; or
  - **passphrase mode**: in `vault-key.age`, encrypted with a passphrase-derived (scrypt) key.
- App-level `app.json` gains `"vault": { "mode": "keychain" | "passphrase", "passphraseCheck":
  "<scrypt-encrypted known token>" }`; the check lets Xenon reject a wrong passphrase before
  touching any vault.

### Vault state machine (per workspace, in memory)
```
NoVault --first secret saved--> Unlocked (identity created, stored per mode)
App start --keychain mode, entry found--> Unlocked
App start --passphrase mode--> Locked --correct passphrase--> Unlocked
Locked --wrong passphrase--> Locked (error VLT-02401)
Any --identity missing / decrypt fails--> Broken --Reset vault (confirmed)--> Unlocked (empty)
Workspace removed --> vault file, key file and keychain entry deleted
```
Mode is chosen when the first vault in the app is created: keychain if a probe write/read/delete
of a test entry succeeds, else passphrase. Once passphrase mode is chosen it stays (a later
"Move keys to keychain" action is out of scope for the MVP).

## Resolution (in memory)

### VariableContext
Built per send/inspect from: local scope (stage 4; empty now) → active environment rows →
collection rows of the request's collection → globals rows; each row's effective value =
current (local or vault) else initial; disabled rows skipped.

### Reference
`{ name, start, end, field }` found by the tokenizer in any text field, using Postman's pattern
`{{([^{}]*?)}}`; the name is taken exactly (no trimming).

### ResolvedReference
`{ name, status: resolved | unresolved | missing_secret | locked_secret | dynamic,
  scope: local | environment | collection | globals | dynamic | none, value (masked if secret),
  secret }`.

### Resolution rules
1. Look up the name through the precedence list; first enabled row wins.
2. If not found and the name starts with `$`, use the dynamic generator if it knows the name
   (fresh value per occurrence); else unresolved. (A user-defined variable named `$x` therefore
   wins over the dynamic `$x` — precedence rule recorded in research.md R3.)
3. Values are resolved by repeated replacement passes over the text while replacements keep
   happening, at most 19 passes (Postman's `VARS_SUBREPLACE_LIMIT`); whatever `{{…}}` remains is
   left literal with status unresolved (this also ends cycles).
4. Unresolved names are left literal. Missing/locked secrets are left literal and reported.

### SendPreflight
Result of checking a request before sending: `{ missing_secrets: [names], locked_secrets:
[names], confirm_environment?: { id, name, color } }`. The send is refused with a specific
error code until the UI re-sends with the matching acknowledgements (contracts/bindings.md).
