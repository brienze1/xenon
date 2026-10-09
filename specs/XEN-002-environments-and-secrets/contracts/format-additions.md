# Contract: workspace format additions (stage 2, still `xenon: 1`)

Fills the slots XEN-001's [workspace-format.md](../../XEN-001-shell-and-storage/contracts/workspace-format.md)
reserved. No existing field changes meaning, so the format version stays `1` (format §10).

## 1. Variable row (replaces format §6.1 for variable tables)

Used by `workspace.yaml` → `variables` (globals), `collection.yaml` → `variables`, and environment
files → `variables`. Rows hold the **initial value** only.

```yaml
# plain row
- key: base_url
  value: "https://dev.acme.test"     # initial value (shared through git)
  enabled: true
  description: ""
# secret row — NO value key, ever
- key: api_key
  secret: true
  enabled: true
  description: "from the vendor portal"
```

Rules:
- `secret` is written only when `true`. A reader that finds `secret: true` **and** a `value`
  ignores the value, never sends it, and Xenon removes it on the next save (with a warning that
  git history may still contain it).
- `key`: case-sensitive, stored exactly as typed, must not contain `{`, `}` or line breaks.
- Order of rows is the order in the editor.
- Folder files (`folder.yaml`) have **no** `variables` key; one found there is preserved as an
  unknown field and ignored.

## 2. Environment file — `environments/<file-name>.yaml`

File name derived from `name` with format §8 rules (unique among environments).

```yaml
xenon: 1
kind: environment
id: 01JA3B4C5D6E7F8G9H0J1K2M3N
name: prod
description: ""
color: red                 # none | green | cyan | magenta | yellow | red
confirmBeforeSend: true
variables:
  - key: base_url
    value: "https://api.acme.com"
    enabled: true
    description: ""
  - key: api_key
    secret: true
    enabled: true
    description: ""
```

Defaults on create: `color: red` and `confirmBeforeSend: true` when the name contains `prod`
(case-insensitive), otherwise `none` / `false`. Same writing rules as format §7 (atomic, unknown
fields preserved, explicit defaults).

## 3. `workspace.yaml` and `collection.yaml`

`variables:` hold rows of §1 (globals and collection variables). No other change.

## 4. What never appears in workspace files

Current values, secret values, the active environment, vault metadata and keychain identifiers.
These live in local app data ([data-model.md](../data-model.md) "Local data").
