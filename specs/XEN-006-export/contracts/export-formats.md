# Contract: export formats

## 1. Postman
Inverse of [XEN-003 postman-mapping.md](../../XEN-003-postman-bruno-import/contracts/postman-mapping.md):
every "Postman → Xenon" row is applied right-to-left, then the item's `postman:` passthrough keys
are added where Xenon has no field. Output: collection v2.1 (`info.schema =
https://schema.getpostman.com/json/collection/v2.1.0/collection.json`), environment and globals
JSON with `_postman_variable_scope`, `_postman_exported_at` (now) and
`_postman_exported_using: "Xenon/<version>"`.

Specific rules:
| Xenon | Postman |
|---|---|
| `auth.type: inherit` | `auth` omitted |
| `auth.type: none` | `{ "type": "noauth" }` |
| `auth.type: unsupported` (source postman) | original `{type, <type>: data}` |
| `auth.type: unsupported` (source bruno) | omitted + summary note |
| oauth2 Xenon-only keys `useToken`, `autoRefresh` | dropped (summary) |
| `scripts.preRequest/postResponse` | `event[]` (`listen: prerequest/test`, `script.exec` = lines, `type: text/javascript`); `scripts.disabled` → `event.disabled: true`; passthrough script metadata merged |
| `inheritScripts: false` | dropped (summary) |
| `body.mode: none` with passthrough original body | original body + `"disabled": true` |
| `srcState` | dropped (path kept in `src`) |
| `examples[]` | `response[]` (`originalRequest` from example request; text body; binary → empty + summary) |
| variable rows | `{ key, value, type (passthrough or "default"/"string"), disabled/enabled, description }`; secret rows → `type: secret`, `value: ""` |
| environment `color`, `confirmBeforeSend` | dropped (summary) |
| ids | `info._postman_id`/`item.id` from passthrough, else the Xenon ULID |

**Folder export**: root = the folder as a v2.1 collection (`info.name` = folder name); the
effective ancestor auth and the ancestor scripts that would run (respecting inherit toggles) are
placed on the exported root (`auth`, `event`); collection variables copied to root `variable`.
**Request export**: same, with one item.
**Workspace export**: zip `<workspace>-postman.zip` containing `collections/<name>.postman_collection.json`,
`environments/<name>.postman_environment.json`, `globals.postman_globals.json`.

## 2. Xenon YAML
- **Workspace**: zip of the workspace folder (format XEN-001 §1) excluding `.xenon/`,
  `.xenon-import-*/`, `*.xenon-tmp`; zip root directory = workspace folder name.
- **Collection / folder (zip)**: zip of its directory (with `collection.yaml`/`folder.yaml`),
  plus `xenon-export.yaml` manifest at the root.
- **Collection / folder (bundle)**: one file:
```yaml
xenon: 1
kind: bundle
root: collection            # collection | folder
exportedAt: "2026-10-09T12:00:00Z"
exportedBy: "Xenon/1.0.0"
item:                       # the collection.yaml / folder.yaml content…
  xenon: 1
  kind: collection
  id: …
  name: Billing API
  …
  children:                 # …plus children in sidebar order (folders with their own children, requests inline)
    - { xenon: 1, kind: request, id: …, name: Create customer, … }
    - { xenon: 1, kind: folder, id: …, name: Invoices, …, children: [ … ] }
secrets: null               # or §3 block
```
- **Request / environment**: its YAML file; when secrets are included, a top-level `secrets:` key
  (§3) is added.
- **Manifest** (`xenon-export.yaml` in zips): `{ xenon: 1, kind: export, root: workspace|collection|folder,
  exportedAt, exportedBy, currentValuesAsInitial: bool, secrets: file|null }`.

## 3. Encrypted secrets
- Plaintext (before encryption):
```json
{ "version": 1,
  "secrets": { "env:<envId>": { "api_key": "…" }, "collection:<id>": { … }, "globals": { … } } }
```
  Only scopes/keys present in the export; ids are the exported items' ids.
- Encrypted with age scrypt recipient from the passphrase (work factor logN 18), ASCII-armored.
- Zips: file `xenon-secrets.age` at the root. Single files: top-level `secrets: |` containing the
  armored block.
- Import: decrypt with the passphrase; map scope ids to the imported items' new ids; write to the
  target vault via XEN-002 multi-store save.

## 4. Fidelity normaliser (tests only)
Applies FR-014 rules before comparing JSON: sort object keys; drop `_exporter_id`,
`_postman_exported_at`, `_postman_exported_using`; treat `disabled: false`, empty `description`,
empty arrays and absent keys as equal; compare `url` by `raw` + `query` + `variable`; blank secret
values.
