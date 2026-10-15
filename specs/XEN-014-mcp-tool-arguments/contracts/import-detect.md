# Contract: import detect (`ImportService.Detect` binding and MCP `import detect|clashes|run`)

The request and response shapes are unchanged. Behaviour changes are below.

## Loose folders

A path to a folder that is **not** one of these layouts is a loose folder:
- a Postman data dump (`archive.json` at the top);
- a Bruno `.bru` or YAML collection;
- a git-native Postman v3 workspace or collection folder;
- a Xenon workspace folder (`workspace.yaml`).

Mixed Bruno formats keep their own error.

For a loose folder:
- Each regular, non-hidden top-level file is read and detected as if its path had been passed on
  its own, in name order. Zip files are extracted and the size limits apply.
- Subfolders and hidden entries are not read. Subfolders are reported as skipped; hidden entries,
  symbolic links and other non-regular entries are ignored silently.
- The folder is first walked as today, which enforces the 200 MB single-file limit and UTF-8 on
  every document file in it. A violation fails the detect with its existing error, for any folder.
- Expansion lists the folder **as passed**, even if Xenon looked into a `postman/` subfolder while
  probing the git-native layout.
- Files that fail to read or detect are skipped.
- When anything was skipped, `warnings` holds one entry:
  `{code: "skipped_part", severity: "warning", message: "Skipped in <folder>: <name> (<reason>), …"}`.
  The reason is `folder` for a subfolder, or the error description without the leading name
  (`not a Xenon, Postman, Bruno or cURL export`, `not valid JSON`, …).
- When no file is importable:

  ```json
  { "error": { "code": "IMP-55415",
    "description": "<folder> has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files" } }
  ```

  Over MCP this is `{"code":"IMP-55415","message":"<same text>"}`.

## Top-level `kind` / `version`

| Inputs (after expansion) | `kind` | `version` |
|---|---|---|
| All the same kind and version | that kind | that version |
| Anything else | `mixed` | `""` |

Examples:
- `[globals.json, full-v2.1.json]` → `mixed` / `""`.
- A loose folder of two v2.1 collections → `postman_collection` / `2.1`.
- `[full-v2.1.json]` → `postman_collection` / `2.1`, as today.

## Window

`sourceLabel('mixed', '')` is `Multiple sources`. A `mixed` preview shows the regular item
table, not the dump assignment.
