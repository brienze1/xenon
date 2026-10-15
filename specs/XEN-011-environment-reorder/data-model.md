# Data Model: Reorder Environments (XEN-011)

## On disk — `workspace.yaml`

| Key | Type | Meaning |
| --- | --- | --- |
| `order` | list of strings | Collections' folder names (unchanged). |
| `environments` | mapping (optional) | Environment-level workspace settings. Absent until the first order-changing operation. |
| `environments.order` | list of strings | Environment ids, first = top of the list. |

Rules (see research R3–R6):

- Absent `environments`, or `environments` not a mapping, or `environments.order` not a list of
  scalars → **no order** (alphabetical). The workspace still loads.
- Duplicated ids → first occurrence wins. Ids with no readable environment file → ignored (kept on
  disk until the next write of the order, which drops them because writes are built from the
  arranged readable ids).
- Unknown keys under `environments:` are preserved by every write.
- Secrets: the list holds only ids — no names, values or secrets (Principle II unaffected).

## Domain — `internal/domain/entity`

### `Workspace` (`workspace.go`) — changed

| Field | Type | Notes |
| --- | --- | --- |
| `EnvironmentOrder` | `[]string` | **New.** `nil` = no stored order; non-nil (possibly empty) = stored order. |

New method: `WithEnvironmentOrder(order []string) Workspace` — returns a copy holding a non-nil
copy of `order`. `LogValue` gains `"environmentOrder": len(w.EnvironmentOrder)`.
`NewWorkspace` leaves it `nil`.

### `EnvironmentList` (`environment_list.go`) — changed

Constructor: `NewEnvironmentList(environments []Environment, order []string, activeId string)`
arranges per research R5.

New methods (all return a fresh `[]string` of ids, never mutate):

| Method | Result |
| --- | --- |
| `IDs()` | readable environment ids in arranged order |
| `OrderMoving(id string, index int)` | `IDs()` without `id`, then `id` inserted at `index` (index ≥ len → last) |
| `OrderAppending(id string)` | `IDs()` + `id` (skips if already present) |
| `OrderReplacing(previous, id string)` | `IDs()` with `previous` swapped for `id` in place (`previous` is present: callers compute this before the replaced file is deleted); appended if `previous` is absent |
| `OrderRemoving(id string)` | `IDs()` without `id` |

Existing helpers (`Contains`, `IsNameTaken`, `CopyName`, `KeepsName`, `IsNameAvailableFor`,
`SiblingFileNames`) are unchanged.

Malformed environments (`Environment.IsMalformed()`) never appear in any id list (their synthetic
`error:<file>` id is not an environment id).

## Persistence model — `internal/integration/persistence/model/workspace.go` — changed

| Field | Type | Notes |
| --- | --- | --- |
| `Environments` | `yaml.Node` with tag `yaml:"environments"` | Raw, so a malformed value never fails decoding. |

- `NewWorkspace(entity)` stores `EnvironmentOrder` (keeps `nil` as "absent").
- `ToEntity()` → `EnvironmentOrder` = scalar items of `environments.order` when well-formed, else
  `nil`.
- `Fields()` → emits `yamlnode.Object("environments", yamlnode.StringList("order", order))`
  right after the `order` field, only when the entity's order is non-nil.

The model needs to carry the non-nil order from `NewWorkspace` to `Fields`: keep it in an
unexported-to-YAML field (`EnvironmentOrder []string yaml:"-"`) set by `NewWorkspace`; `Fields`
reads that field. `ToEntity` reads the raw node.

## DTOs — `internal/integration/entrypoint/dto`

### `ReorderEnvironment` (new file `reorder_environment.go`)

| JSON | Go | Validation |
| --- | --- | --- |
| `local_id` | `LocalId string` | `required,not_blank` |
| `id` | `Id string` | `required,not_blank` |
| `index` | `Index int` | `min=0` |

Response: existing `dto.EnvironmentList` (`NewEnvironmentList(entity.EnvironmentList)`), already
ordered as the entity is.

### `MCPEnvEdit` (`mcp_env_edit.go`) — changed

New field `Index *int json:"index"` and method `ReorderEnvironment(localId string)
ReorderEnvironment` (nil index → `-1`, which fails `min=0` validation with `ENV-18400`).

## State transitions of the stored order

```text
           create/duplicate/import/reorder
  (absent) ───────────────────────────────▶ [ids…]
     ▲                                         │ reorder/create/duplicate/import/replace → rewritten from IDs()
     │ (never: nothing removes the key)        │ delete of a listed id → id removed
     └─────────────────────────────────────────┘ rename/update/list → untouched
```
