# Data Model: Alias Environments

## RequestAlias (entity, `internal/domain/entity/request_alias.go`) — changed

| Field | Type | Notes |
| --- | --- | --- |
| ID | string | unchanged |
| Name | string | unchanged |
| Version | Request | unchanged |
| **EnvironmentIDs** | []string | NEW. Empty/nil = All environments. Normalized (R6) on every write. |

Methods:

- `NewRequestAlias(id, name string, of Request, environmentIds []string)`: stores
  `NormalizedEnvironmentIDs(environmentIds)`.
- `IsLinked() bool`: `len(EnvironmentIDs) > 0`.
- `IsAvailableIn(active *Environment) bool`: `!IsLinked() || (active != nil && slices.Contains(EnvironmentIDs, active.ID))`.
- `WithEnvironments(ids []string) RequestAlias`: returns a copy with normalized ids.
- `Edited` / `Renamed`: unchanged; they keep `EnvironmentIDs`.

Default (the request itself) has no environments field and is always available (FR-004).

## Pure helpers (`internal/domain/entity/alias_environments.go`) — new

- `NormalizedEnvironmentIDs(ids []string) []string`: trims each id, drops blanks, dedupes in
  first-seen order, and returns nil when empty.
- `AliasEnvironmentLabels(ids []string, environments []Environment) []string`: the name per id, or
  `"Deleted environment"` when no environment has that id.

## HistoryEnvironment (entity, new) and HistoryEntry (changed)

| Field | Type | Notes |
| --- | --- | --- |
| ID | string | environment id at send time |
| Name | string | name at send time |
| Color | enums.EnvColor | colour at send time |

`HistoryEntry.Environment *HistoryEnvironment` (nil = no active environment, or an entry from
before this feature). `HistoryEntry.InEnvironment(active *Environment) HistoryEntry` sets it from
the active environment, or leaves nil.

## YAML (request file) — `aliases[]` item

```yaml
aliases:
  - id: 01J…
    name: mx-test-user
    environments:        # NEW, written only when non-empty
      - 01HML…
      - 01DEV…
    method: POST
    url: '{{api}}/login'
    …
```

Model `model.RequestAlias.Environments []string \`yaml:"environments"\``. In `Node()` the field
comes right after `name` and is emitted only when non-empty. When reading, a missing key means nil,
which means All environments.

## History JSON — send entry

```json
{ "id": "…", "alias_id": "…", "alias_name": "mx-test-user",
  "environment": { "id": "01HML…", "name": "HML", "color": "green" } }
```

`environment` is `omitempty`; `historyVersion` stays `1`, because the addition is backwards
compatible.

## State transitions — alias availability

```
               SetEnvironments([])                 SetEnvironments([HML,…])
  All environments  ◀──────────────────────────────▶  Linked {HML,…}
  (always available)                                   available ⇔ active ∈ set
                                                       env deleted → id stays (Deleted environment)
                                                       all ids deleted → unavailable everywhere
```

## Validation rules

| Rule | Where | Code |
| --- | --- | --- |
| alias exists | `usecase.ResolveAlias` | `ALS-01404` |
| request not changed on disk since opened | `manageAlias.load` | `REQ-02409` |
| every id is an environment of the workspace (write only) | `manageAlias.checkEnvironments` | `ALS-10400` |
| `SetAliasEnvironments` body: `local_id`, `request_id`, `alias_id` required and not blank | `validateSetAliasEnvironments` | `ALS-11400` |
| alias available for the active environment (send) | `sendRequest.Send` | `ALS-09403` |
