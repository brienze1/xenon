# Contract: MCP tool input schemas (`tools/list`)

Each tool in `tools/list` keeps its `name`, `description` and `annotations`. Its `inputSchema`
becomes the following.

## Shape

```json
{
  "type": "object",
  "properties": {
    "<selector>": { "type": "string", "enum": ["<action>", "..."] },
    "<argument>": {
      "type": "<json type>" | ["<json type>", "..."],
      "description": "<sentence>[ Used by <actions>.][ Required for <actions>.]",
      "items": { },
      "properties": { },
      "additionalProperties": { },
      "enum": ["..."]
    }
  },
  "required": ["<selector>", "<always-required argument>"]
}
```

Rules:
- `<selector>` is `action`, or `target` for `delete`. It is required when the tool has more than
  one action, as today.
- `items`, nested `properties`, `additionalProperties` and `enum` appear only when they apply.
- `required` is omitted when empty.
- The top level never has `oneOf`, `anyOf` or `allOf`.
- Every top-level argument has a `type` and a non-empty `description`.
- "Used by" is present only on tools with more than one action.
- Arguments are never coerced: a client must send `paths` as an array, `depth` as an integer or
  `"all"`, and so on.

## `workspace_edit` (exact)

```json
{
  "type": "object",
  "properties": {
    "action": { "type": "string", "enum": ["create", "open", "switch", "rename"] },
    "id":   { "type": "string", "description": "Workspace id, from workspace list. Used by switch and rename. Required for switch and rename." },
    "name": { "type": "string", "description": "Workspace name, at most 200 characters. Used by create and rename. Required for create and rename." },
    "path": { "type": "string", "description": "Absolute path of the workspace folder. Used by create and open. Required for create and open." }
  },
  "required": ["action"]
}
```

## `import` (exact)

```json
{
  "type": "object",
  "properties": {
    "action":     { "type": "string", "enum": ["detect", "clashes", "run", "discard"] },
    "paths":      { "type": "array", "items": { "type": "string" }, "description": "Absolute paths of the files or folders to import. Used by detect, clashes and run. Required for detect, clashes and run." },
    "source_id":  { "type": "string", "description": "Source id returned by detect; omitted means nothing to discard. Used by discard." },
    "on_clash":   { "type": "string", "enum": ["keep_both", "replace", "skip"], "description": "How to settle a name clash; default keep_both on run. Used by clashes and run." },
    "passphrase": { "type": "string", "description": "Passphrase of an export with encrypted secrets. Used by clashes and run." }
  },
  "required": ["action"]
}
```

`import` decodes one DTO (`MCPImport`) for all four actions. Each field narrows its "Used by"
with the `used_by` tag (data-model.md).

## `tree` (`depth` excerpt)

```json
"depth": { "type": ["integer", "string"], "description": "Levels to list: an integer of 1 or more, or \"all\"; default 1. Used by list." }
```

## `settings_edit`

`ai_access` is not declared. Sending it is still refused with the human-only error, as today.

## Wording of "Used by" / "Required for"

Actions are listed in the tool's action order and joined as `a`, `a and b`, or `a, b and c`.
Each list ends with a period.
