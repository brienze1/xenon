# Import fixture corpus (XEN-003)

| Folder | What | How it was made |
|---|---|---|
| `postman/full-v2.1.json` | collection covering every row of `contracts/postman-mapping.md` | hand-authored against the Postman v2.1.0 schema |
| `postman/full-v2.0.json` | the same collection with v2.0 auth maps and schema | derived from the v2.1 file |
| `postman/dev.postman_environment.json`, `postman/globals.json` | environment (plain, disabled, secret with/without value) and globals | hand-authored |
| `postman/dump/`, `postman/dump.zip` | data dump: `archive.json`, 3 collections, 2 environments, one unknown part (`mock/`) | hand-authored; the zip holds the same files |
| `postman/v1.json`, `postman/v3/` | refused formats (v1 JSON, v3 Native Git YAML) | hand-authored |
| `bruno/` | `.bru`, OpenCollection YAML and a mixed collection | hand-authored, see `bruno/README.md` |
| `curl/corpus.yaml` | 32 commands, each in bash, Windows cmd and PowerShell form, with the expected request | hand-authored |
| `misc/` | random/malformed/non-UTF-8 JSON, other YAML, unsafe zips (`..`, absolute, symlink) and a nested dump zip | hand-authored |

All secret-looking values are fake.
