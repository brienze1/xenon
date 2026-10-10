# Bruno fixtures (hand-authored)

Bruno 3.1 was not available when these fixtures were created, so both collections were written
by hand instead of being exported by Bruno:

- `bru-collection/` follows the `bruno-lang` v2 grammar (blocks `meta`, method blocks incl.
  `http`, `params:query|path`, `headers`, `auth:*`, `body:*`, `vars:pre-request|post-response`,
  `assert`, `script:pre-request|post-response`, `tests`, `docs`, `settings`; `~` disables a row,
  `@file(...)` / `@contentType(...)` annotate multipart rows) as described in
  `specs/XEN-003-postman-bruno-import/research.md` (R5) and `contracts/bruno-mapping.md`.
- `yaml-collection/` follows the OpenCollection layout from the same documents
  (`opencollection.yml`, `folder.yml`, request `*.yml` with `info`, `http`, `runtime`,
  `settings`, `docs`, and `environments/*.yml`).

Both describe the same "Bruno Shop" collection. Replace them with real Bruno 3.1 exports when
available and keep the scenario expectations in `test/integration/features/import/bruno_*.feature`
in sync.
