# Research: Environments, Variables and Secrets (Stage 2)

Verified 2026-10-09 against Postman docs, the `postman-collection` SDK source, and module proxy
versions.

## R1. Postman variable semantics (parity target)
- **Decision**: precedence local > environment > collection > globals (Postman: local > data >
  environment > collection > global; data scope is out of MVP); disabled rows fall through to
  the next scope; unresolved names sent literally; names matched exactly with Postman's token
  pattern `{{([^{}]*?)}}` (no trimming, case-sensitive); nested resolution by repeated passes up
  to 19 (`VARS_SUBREPLACE_LIMIT`); values are strings.
- **Initial/current**: sends and scripts use current; Persist copies current → initial; Reset
  copies initial → current; exports carry initial values only. Xenon models "no current value"
  as the absence of a local override (so editing the initial value is reflected immediately when
  there is no override) — same visible behaviour as Postman's pre-filled current value.
- **Confirm before send**: not a Postman feature; it is a PRD addition. Enforced in the backend.
- **Postman environment JSON** (`values[{key,value,type,enabled}]`, `type: default|secret`,
  `_postman_variable_scope`) is the stage-3 import target; collections use `disabled` instead of
  `enabled` — noted for stage 3.

## R2. Dynamic variables
- **Decision**: all 118 names ([contracts/dynamic-variables.md](contracts/dynamic-variables.md)),
  fresh value per occurrence (SDK calls the generator per match), checked after user scopes (SDK
  appends dynamic variables as the last variable list). Generator =
  `github.com/brianvoe/gofakeit/v7@v7.17.1` (MIT) + in-repo lists for the ~40% with no direct
  equivalent. Behind an application adapter `DynamicValueGenerator` so BDD can seed it.
- **Alternatives**: a subset (rejected by the user); porting faker.js word lists (large, no value).

## R3. Encryption — age
- **Decision**: `filippo.io/age@v1.3.2`. Each workspace vault is encrypted to its own X25519
  identity (`age.GenerateX25519Identity`, stored as `AGE-SECRET-KEY-1…` text, ~74 bytes), binary
  (not armored) output, written with XEN-001's atomic writer.
- **Passphrase mode**: the identity text is wrapped in `vault-key.age` with
  `age.NewScryptRecipient(passphrase)` (work factor logN 18 ≈ 1 s; property
  `XENON_SCRYPT_LOG_N` lowers it to 10 in BDD only). `passphraseCheck` in `app.json` is a known
  token encrypted the same way, used to validate the passphrase once per start.
- **Rationale**: one identity per workspace keeps workspaces cryptographically independent;
  switching storage mode means re-wrapping 74 bytes, never re-encrypting vaults.
- **Alternatives**: post-quantum hybrid identities (unnecessary, 2 KB recipients); NaCl secretbox
  with own KDF (reinventing age).

## R4. OS keychain — go-keyring
- **Decision**: `github.com/zalando/go-keyring@v0.2.8` (MIT, no cgo): macOS via
  `/usr/bin/security`, Windows Credential Manager (≤ 2,560 bytes — identity is 74), Linux Secret
  Service over D-Bus. Service `dev.brienze.xenon.vault`, user = workspace local id.
- **Availability probe** at first vault creation: Set/Get/Delete of
  `dev.brienze.xenon.vault/probe`. `ErrNotFound` = available; `ErrUnsupportedPlatform` or any
  other error (e.g., no D-Bus session) = unavailable → passphrase mode.
- **Known limitation (documented, accepted)**: on macOS the item is trusted for
  `/usr/bin/security`, so another process running as the same user can read it without a prompt.
  Same-user processes can already read the user's files, so this does not lower the threat model
  below "secrets are safe from git and other users"; it means no keychain prompts for Luis.
- **Testing**: `keyring.MockInit()` / `MockInitWithError` behind an injected `Keychain` adapter
  (setter), so BDD covers both modes.

## R5. Secrets in memory, clipboard
- **Decision**: no memory-zeroing promises (Go strings are immutable and copied). Secrets are
  decrypted per operation and kept in the vault service's memory while unlocked. `CopySecret`
  writes the clipboard from Go (Wails `app.Clipboard.SetText`) so values never cross the bridge
  for a copy; the UI only gets a secret value from an explicit Reveal.
- **Alternatives**: clearing the clipboard after N seconds (rejected for MVP: surprising, and
  clipboard managers keep history anyway).

## R6. Highlighting and autocomplete in the UI
- **Decision**: the URL bar becomes a single-line CodeMirror 6 editor (shared `code-editor` in a
  `singleLine` variant) so decorations and completion use the same mechanism everywhere; key/value
  table cells use the same single-line editor. `shared/ui/code-editor` stays domain-free: it
  accepts generic `decorations: {from, to, tone, tooltip}` and `completionSource(prefix)`
  callbacks; `features/request-builder` and `features/environments` feed them from
  `VariableService.Inspect/Suggest` (debounced 120 ms).
- **Alternatives**: overlay `<input>` highlighting (fragile caret/scroll sync).

## R7. Where resolution plugs in
- **Decision**: XEN-001's send pipeline becomes: resolve effective request → **preflight**
  (confirm-before-send, missing/locked secrets) → **resolve variables** on a copy (every text
  field per FR-012, including inherited auth/headers and the binary file path) → auto headers
  computed on the resolved request → send → history (stores the unresolved request). The same
  resolver powers `Inspect` (no dynamic generation; reports `dynamic` status instead).
