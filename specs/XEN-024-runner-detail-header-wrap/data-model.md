# Data Model: Runner Row Detail Keeps Header Names Readable

No data changes. No stored file, run format, DTO, mapper or Wails binding changes. The feature
only changes how two existing shapes are laid out.

## Displayed shapes (unchanged)

### RunSentRequest.headers (`frontend/src/platform/mappers/runner.ts`)

| Field | Type | Notes |
|---|---|---|
| `key` | string | Header name as sent; shown as the list term (`dt`) |
| `value` | string | Header value as sent, secrets already masked by Go; shown as the definition (`dd`) |

Order: as sent. Empty array → no list rendered.

### ApiParameter (`frontend/src/platform/mappers/documentation.ts`)

| Field | Type | Notes |
|---|---|---|
| `name` | string | Shown in the Name cell, inside the capped name block |
| `required` | boolean | `required` marker after the name, same line |
| `in`, `description`, `example`, `variable` | — | Unchanged cells |

## Validation rules

None added. Values are rendered verbatim as today (FR-008).
