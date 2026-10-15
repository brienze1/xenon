---
description: "Task list for XEN-009 — Import Postman v3 collections, environments and git-native folders"
---

# Tasks: Import Postman v3 Collections, Environments and Git-Native Folders

**Input**: `specs/XEN-009-postman-v3-import/` (spec, plan, research, data-model,
contracts/v3-to-v21-mapping.md, contracts/import-service.md, quickstart)

**Prerequisites**: XEN-001 … XEN-008 merged on `main` (`6f61014`).

"R#" refers to the numbered decisions in `research.md`.

**Tests**: MANDATORY (constitution V).

- Feature files are tagged `@all @xen_009 @import @<area>`. `/create-tests` writes them and proves them red, and
  `/implement-task` turns them green.
- Only the outer boundaries are faked: file-system roots, the keychain and the clock.
- They use only existing steps (plan, BDD Coverage).

**Stories**:

- US1: v3 collection JSON (P1)
- US2: v3 git-native collection folder and single request file (P1)
- US3: v3 environments and globals (P1)
- US4: whole `postman workspace pull` folder (P2)
- US5: MCP `import` tool parity (P2)

Every Go task follows its skill in `.claude/skills/` (entities, enums, usecases, services, adapters, webservices,
dtos, persistence, models, dependency-injection, controllers, naming-conventions, comments). Every frontend task
follows `frontend-architecture`. Run `make sort-imports` after each batch of Go changes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: the user story the task belongs to

---

## Phase 1: Setup (fixtures, PRD)

**Purpose**: Build the fixture corpus every story's scenarios read, and put v3 in the PRD scope (FR-026).

- [ ] T001 Write `test/integration/resources/import/postman/v3/rich-v2.1.json`. It is a Postman v2.1 collection
  "Rich API" in the canonical form of research R14:
  - ids on the collection (`_postman_id`), every folder, request and example;
  - `request.url` as `{raw, query, variable}`;
  - auth rows with `type: "string"`;
  - examples with `_postman_previewlanguage`;
  - events without ids.

  It covers:
  - folders `Users` (with basic auth, a description and a pre-request script) and `Bodies`;
  - requests:
    - `Get user`: GET with disabled header and query rows with descriptions, path variable `id` with a
      description, a request description, a test script, and a `200 OK` example;
    - `Inherit auth` (no auth);
    - `No auth` (`noauth`);
    - `JSON`, `Text`, `XML`, `HTML`, `JavaScript` raw bodies;
    - `Form` (formdata with text, disabled text and a file `src` that does not exist);
    - `Urlenc` (with a disabled row);
    - `GraphQL` (query + variables);
    - `Binary` (file `src` that does not exist);
    - `ApiKey`, `Digest`, `Hawk` auths;
    - `Redirects` (`protocolProfileBehavior {followRedirects:false, strictSSL:false, maxRedirects:3}`);
  - collection bearer auth `{{token}}`, collection pre-request and test scripts, collection variables
    `baseUrl`, a disabled `off`, and a description.

  Values are fake. The collection has no `type: secret` variable.
- [ ] T002 Generate `test/integration/resources/import/postman/v3/rich/` with
  `postman collection migrate test/integration/resources/import/postman/v3/rich-v2.1.json -o <tmp>`. Copy the output
  folder in unchanged. No `rich` request name contains a character Windows forbids (T001's names are plain words),
  so `migrate`'s file names are portable as they are (constitution VII).
- [ ] T003 Generate `test/integration/resources/import/postman/v3/rich.json`, the same collection in
  `postman collection get --json` shape (research R1):
  - top-level `{id, name, $kind: "collection", description, variables, auth: [ {id, name, type, credentials: [{key, value}]} ], scripts, items}`;
  - nested folder objects with `$kind: "collection"`;
  - request objects with list-form `headers`, `queryParams`, `pathVariables` and inline `examples`;
  - `settings: {disabledSystemHeaders: []}` on every request (as real CLI output has), except `Redirects`, which
    keeps its settings;
  - no `order` keys.

  Write the generator as a short Python script documented in the README (T010). Run it on `rich/` and commit only
  its output.
- [ ] T004 [P] Hand-write `test/integration/resources/import/postman/v3/extras/` (R14), a v3 collection
  folder "Extras". It holds:
  - `.resources/definition.yaml` with:
    - a 2-entry multi-auth list (bearer, then basic);
    - a `graphql:beforeQuery` script;
    - a `secret: true` variable `apiSecret` = `s3cr3t-fixture-collection`;
  - folder `Nested/.resources/definition.yaml` with `variables` and `order: 1000`;
  - `Map headers.request.yaml` (headers and auth credentials as YAML maps, `order: 2000`);
  - `List headers.request.yaml` (the same request with list forms, `order: 2000`, so the tie is broken by name);
  - `No order.request.yaml` (no `order`);
  - `Health-check.request.yaml` with `name: 'Health/check'`;
  - `Query.request.yaml` with `$kind: graphql-request`, `query`, `variables` and a `beforeQuery` script;
  - `Stream.request.yaml` with `$kind: grpc-request`;
  - `Socket.request.yaml` with `$kind: websocket-request`;
  - `Future.request.yaml` with `$kind: carrier-pigeon-request`;
  - `Redirects.request.yaml` with `settings.maxRedirects: 3`;
  - `.resources/Map headers.resources/examples/` with two `*.example.yaml` files with `order` 2000 and 1000.
- [ ] T005 [P] Hand-write `test/integration/resources/import/postman/v3/env/`:
  - `staging.environment.yaml`:
    - plain `host`;
    - disabled `debug`;
    - `apiKey` with `type: secret` and value `s3cr3t-fixture-env`;
    - `token` with `secret: true` and value `s3cr3t-fixture-token`;
    - `linked` with `secret: true` and `source: {provider: postman, postman: {type: cloud, secretId: 11111111-2222-3333-4444-555555555555, vaultId: 66666666-7777-8888-9999-000000000000}}` and no value;
  - `noname.environment.yaml` (no `name`);
  - `cli-environment.json` (CLI JSON shape with `id, owner, color: null, createdAt, updatedAt, lastRevision, lastUpdatedBy, team: null` and a secret whose value is `********`);
  - `workspace.globals.yaml` (`g1` plain, `gsecret` `type: secret` = `s3cr3t-fixture-global`);
  - `cloud-globals.json` (top-level `workspace` field);
  - `local-globals.json` (`id: "postman/globals/workspace.globals.yaml"`, `path`).
- [ ] T006 [P] Hand-write `test/integration/resources/import/postman/v3/workspace/` (R14):
  - `postman/collections/Billing/` (2 requests) and `postman/collections/Users/` (copy of `extras/` minus the
    grpc and websocket files);
  - `postman/environments/staging.environment.yaml` and `prod.environment.yaml`;
  - `postman/globals/workspace.globals.yaml`;
  - `postman/specs/openapi.yaml`, `postman/flows/f.json`;
  - `.postman/resources.yaml`;
  - `tools/latin1.json` (non-UTF-8 bytes, copied from `misc/latin1.json`).

  Also write `workspace.zip` holding the same tree.
- [ ] T007 [P] Generate `test/integration/resources/import/postman/v3/large-1000-v3.zip` by zipping the output of
  `postman collection migrate test/integration/resources/import/postman/large-1000.json` (1,011 files).
- [ ] T008 [P] Hand-write `test/integration/resources/import/postman/v3/malformed/`: a collection folder with
  `.resources/definition.yaml` and `Broken.request.yaml`. The broken file has `headers` as a string
  `'s3cr3t-fixture-broken'` where a list or map is expected. That is a yaml type error whose raw message would quote
  the value.
- [ ] T009 Update `docs/prd/xenon-mvp-prd.md` (FR-026):
  - in "What can be imported", change the Postman bullet to "Postman collection v2.1 JSON and v3 (the JSON
    `postman collection get` prints, git-native `postman/collections` folders, single `*.request.yaml`), with its
    scripts, auth, variables and folder order";
  - change the environment bullet to "Postman environment and globals JSON, and v3 environment and globals YAML";
  - add "A `postman workspace pull` folder (collections, environments and globals, each assigned to a workspace)";
  - in Fidelity, replace "Postman v3 YAML is not in the MVP." with "Postman v3 (JSON, git-native YAML and
    workspace folders) is imported through the v2.1 mapping (XEN-009); export stays v2.1."
- [ ] T010 [P] Add a "Postman v3 (XEN-009)" section to `test/integration/resources/import/README.md`. It has:
  - a row per v3 fixture folder or file, saying how it was made (`collection migrate` / generator script /
    hand-written);
  - the generator script used in T003;
  - the hand-added cases list from T004.

  Change the existing `postman/v3/` row from "refused formats" to "`postman/v3/native-git/`: detected as a v3
  collection folder and a v3 request". The two original files moved there because `postman/v3/` now holds every
  v3 fixture.

---

## Phase 2: Foundational (blocks every story)

**Purpose**: Shared domain types, ports and plumbing every story uses. No behaviour visible yet.

- [ ] T011 [P] Add `PostmanRequest = "postman_request"` and `PostmanWorkspace = "postman_workspace"` to
  `internal/domain/enums/importkind/value.go`.
- [ ] T012 [P] Add `ProtocolScript = "protocol_script"` and `LinkedSecret = "linked_secret"` to
  `internal/domain/enums/warningcode/value.go`. Both are warning severity: the default, not in the info list of
  `entity.NewImportWarning`.
- [ ] T013 [P] Create `internal/domain/entity/import_collection.go` with `ImportCollection{Root ImportNode; Warnings []ImportWarning; Secrets map[string]string}`
  and constructor `NewImportCollection(root, warnings, secrets)`. A nil `secrets` becomes an empty map.
- [ ] T014 [P] Create `internal/domain/entity/postman_v3_source.go` with `PostmanV3Source{Collections []ImportCollection; Request *ImportCollection; Environments []ImportVariables; Globals *ImportVariables; Warnings []ImportWarning; Skipped []PostmanDumpPart}`.
  Add `SkippedWarnings()`, which returns one `warningcode.SkippedPart` warning per part with message
  `The workspace's <part> part (<n> file|files) is not a collection, an environment or globals and was skipped`,
  empty item id and path = part name. Reuse `fileNoun` from `postman_dump.go`.
- [ ] T015 [P] Add `IsPostmanV3() bool` (`Version == "3"`) to `internal/domain/entity/import_input.go`, with a
  constant `PostmanV3Version = "3"` in the same package.
- [ ] T016 Change `usecase.ReadPostmanCollection` in `internal/application/usecase/read_postman_collection.go` to
  `ReadCollection(ctx, entity.ImportDocument) (*entity.ImportCollection, error)`. Update
  `internal/integration/webservice/postman.go` to return
  `entity.NewImportCollection(root, conversion.Warnings(), secrets)`, where `secrets` comes from T017. Update
  `internal/application/service/import_readers.go` `collection()`, which reads `ImportCollection` and calls
  `read.addCollection` (T018).
- [ ] T017 R7, v2.1 secret collection variables:
  - in `internal/integration/webservice/postman/dto/key_value.go`, add `IsSecret()` (`Type == "secret"`) and make
    `ToVariableRow` set `row.Secret = true` and keep `type` out of the passthrough for secret rows;
  - in `internal/integration/webservice/postman/dto/collection.go`, `ToEntity` collects `key → value` for secret
    rows with non-empty values into a secrets map, then blanks the row's `InitialValue`;
  - change `ToEntity` to return `(entity.ImportNode, map[string]string)`;
  - update its callers (`postman.go`) and the unit tests in the package.
- [ ] T018 Add `addCollection(kind enums.ImportItemKind, collection entity.ImportCollection)` to
  `internal/application/service/import_read.go`. It behaves like `addTree` and also sets
  `item.CollectionSecrets = collection.Secrets` when the map is non-empty. `importWriter.keepCollectionSecrets`
  (unchanged) then stores them.
- [ ] T019 Create the port `internal/application/usecase/read_postman_v3.go`:
  `type ReadPostmanV3 interface { ReadPostmanV3(ctx context.Context, input entity.ImportInput) (*entity.PostmanV3Source, error) }`.
- [ ] T020 Wire the port through the code:
  - `internal/integration/adapter/postman_webservice.go` embeds `usecase.ReadPostmanV3`;
  - `internal/application/service/manage_import.go` `NewManageImportService` takes `readPostmanV3`;
  - `importReaders` gets field `readPostmanV3`;
  - `internal/infra/dependency/application/manage_import.go` passes `integration.GetPostmanWebservice()`;
  - create `internal/integration/webservice/postman_v3.go` with
    `func (p *postmanWebservice) ReadPostmanV3(ctx, input) (*entity.PostmanV3Source, error)` logging
    `Started`/`Finished` (counts only, never values) and dispatching on `input.Kind` to the per-story readers
    added below.
- [ ] T021 In `internal/application/service/import_readers.go` `read()`, route `input.IsPostmanV3()` before the kind
  switch to a new `postmanV3(ctx, input, read)`. It calls `readPostmanV3`, then:
  - for each `Collections` entry, `read.addCollection(importitemkind.Collection, …)`;
  - for `Request`, `read.addCollection(importitemkind.Request, …)`;
  - for each `Environments` entry, an `importitemkind.Environment` variables item;
  - for `Globals`, an `importitemkind.Globals` variables item;
  - appends `Warnings` and `SkippedWarnings()` to `read.warnings`.
- [ ] T022 Create the v3 DTO package `internal/integration/webservice/postmanv3/dto/` with shared contract types
  (contracts/v3-to-v21-mapping.md, "Normalisation applied first"). Every type decodes from both JSON and YAML:
  - `key_values.go`: `KeyValues` decodes a list of `{key, value, description, disabled}` or a map, keeping map
    order through `yaml.Node` for YAML and decoder token order for JSON. Null values become `""`. Unknown keys go
    into a raw map;
  - `credentials.go`;
  - `script.go`;
  - `variable.go`;
  - `conversion.go`: collects the R8 warnings with `entity.JoinImportPath` paths;
  - `postman_v21.go`: small builders returning `map[string]any` v2.1 fragments, created through constructors per
    the dtos skill.

  Add unit tests for list vs map decoding in `key_values_test.go`.
- [ ] T023 Run `make sort-imports` and `go build ./...`. The existing suite stays green: `make bdd TAG=@import` and
  `go test ./internal/...`.

**Checkpoint**: The plumbing compiles. v3 sources are still detected as before; the stories change detection.

---

## Phase 3: User Story 1 — v3 collection JSON (P1) 🎯 MVP

**Goal**: A `postman collection get --json` file (or pasted JSON) imports with v2.1 fidelity.

**Independent test**: `rich.json` and `rich-v2.1.json` imported into two workspaces give equal collection folders
apart from ids.

### Tests (write first, prove red)

- [ ] T024 [US1] Create `test/integration/features/import/postman_v3_collection.feature` (`@xen_009 @import
  @postman_v3_collection`) with the JSON scenarios from the plan's BDD Coverage:
  - (1) JSON vs v2.1 folder equality, using `the folder "…" should equal the folder "…" apart from ids` on the two
    workspaces' `collections/rich-api`;
  - (2) bodies per type, with disabled rows and the missing-file `missing_file` warnings for `Form` and `Binary`;
  - (3) auth: collection bearer, folder basic, `Inherit auth` → `auth.type` inherit, `No auth` → none, `Hawk` →
    `unsupported_auth` warning;
  - (4) scripts at three levels;
  - (5) inline examples;
  - (12) `Redirects` keeps `maxRedirects` in its `postman:` block and maps `followRedirects`/`strictSSL`. A
    request with `settings: {disabledSystemHeaders: []}` gets no `postman.protocolProfileBehavior`;
  - (15) clash keep_both / replace / skip on re-import of `rich.json`;
  - Detect on pasted `rich.json` text gives `postman_collection`/`3`.

  Run `make bdd TAG=@postman_v3_collection` and record that it is red.
- [ ] T025 [P] [US1] Add the v2.1 secret-collection-variable scenario to
  `test/integration/features/import/postman_collection.feature` (R7). A docstring collection is written with
  `the file "…" exists with content`, with a variable `{key: "apiSecret", value: "s3cr3t-fixture-v21", type: "secret"}`.
  After import:
  - the collection file has `variables.N.secret` = `true` and no value;
  - `no file under the workspace folder should contain "s3cr3t-fixture-v21"`;
  - `VaultService.Reveal` returns the value for the collection scope.

  Prove it red.

### Implementation

- [ ] T026 [US1] Detection of v3 JSON (R4.4):
  - `internal/integration/persistence/model/import_probe.go`: add `Kind string` (`json:"$kind" yaml:"$kind"`),
    `Values []any` (`json:"values" yaml:"values"`), `Workspace any` (`json:"workspace" yaml:"workspace"`), `ID any`
    (`json:"id" yaml:"id"`) and a `Name` field if missing. `ToEntity` fills `PostmanKind`, `HasValues`,
    `HasWorkspaceField` and `DocumentID`;
  - `internal/domain/entity/import_probe.go`: add the fields and `IsPostmanV3Collection()` (no v2 schema and
    `PostmanKind == "collection"`);
  - `internal/application/usecase/detect_import_format.go` `detectJson`: return `importkind.PostmanCollection`,
    `entity.PostmanV3Version` for it, before the environment checks;
  - add rows to `internal/application/usecase/import_detection_test.go`.
- [ ] T027 [US1] Request, folder, body, auth and example DTOs in
  `internal/integration/webservice/postmanv3/dto/`:
  - `collection.go` (top-level and folder `$kind: collection`, with `items` decoded polymorphically by `$kind`);
  - `item.go`;
  - `request.go` (`http-request` and `graphql-request`);
  - `body.go`;
  - `auth.go` (a single object or a list; a list with 2+ entries takes the first, warns `unsupported_auth` with the
    R8 message, and keeps the list under `postmanV3.auth`);
  - `example.go`.

  Each has a `ToPostmanV21(conversion)` that builds exactly the v2.1 shape in contracts/v3-to-v21-mapping.md:
  - non-HTTP `$kind`s, and any unknown `$kind` ending in `-request`, are dropped with the
    `unsupported_request_type` warning; any other unknown `$kind` raises `IMP-11400`;
  - `graphql:*`/`grpc:*` collection and folder scripts are dropped from `event` with a `protocol_script` warning and
    kept under `postmanV3.scripts`;
  - unknown keys are copied through;
  - empty `settings` values are dropped, and `protocolProfileBehavior` is omitted when nothing is left (contract
    normalisation);
  - a secret collection variable whose value is `********` gets `""`.

  Unit-test the builder outputs against small golden maps in `collection_test.go`.
- [ ] T028 [US1] In `internal/integration/webservice/postman_v3.go`, for `PostmanCollection` with a single JSON
  document:
  1. decode `dto.Collection`. On error, raise `IMP-11400` `"<path> is not a valid Postman v3 file: <reason>"`,
     where `<reason>` is one of the exact R9 forms, built by a shared `postmanv3/dto` helper
     `DecodeError(path, err)`. Unit-test it with a yaml type error carrying
     `s3cr3t-fixture` and assert the value is absent;
  2. `ToPostmanV21`;
  3. `json.Marshal`;
  4. run the existing v2.1 path (`json.Unmarshal` into `postman/dto.Collection`, then
     `ToEntity(dto.NewConversion(idutils.NewUlid))`);
  5. return one `ImportCollection` whose warnings are the v2.1 warnings followed by the v3 conversion warnings.

  Factor the shared v2.1 step out of `postman.go` into an unexported helper both use.
- [ ] T029 [US1] Turn the US1 scenarios and T025 green. Run `make bdd TAG=@postman_v3_collection`,
  `make bdd TAG=@postman_collection` and `make bdd TAG=@import`.

**Checkpoint**: v3 JSON imports. The MVP is usable from `postman collection get --json`.

---

## Phase 4: User Story 2 — v3 git-native collection folder and single request (P1)

**Goal**: A `postman/collections/<name>/` folder (or zip), or a single `*.request.yaml`, imports.

**Independent test**: `rich/` imports equal to `rich-v2.1.json` apart from ids. `extras/` cases import as
specified.

### Tests (write first, prove red)

- [ ] T030 [US2] Extend `test/integration/features/import/postman_v3_collection.feature` with the folder scenarios:
  - (1b) `rich/` vs `rich-v2.1.json` folder equality;
  - (5b) example files ordered by `order`;
  - (6) order: `List headers` before `Map headers` (tie, by name), and `No order` last;
  - (7) `Health/check` name from `name:`;
  - (8) multi-auth first-entry + `unsupported_auth` warning naming `basic`;
  - (9) `protocol_script` warning for `graphql:beforeQuery`;
  - (10) `Stream` / `Socket` `unsupported_request_type` warnings with the exact R8 message, and `Query` imported
    as POST with a GraphQL body and a pre-request script;
  - (11) `folder_variables` warning on `Nested`;
  - (13) `extras/Map headers.request.yaml` imported into a folder of an existing collection with its two examples,
    and as top-level refused `IMP-35400`;
  - (14) `malformed/` → `IMP-11400` with the exact description
    `Broken.request.yaml is not a valid Postman v3 file: line 4: a value has the wrong type`, so
    `s3cr3t-fixture-broken` is absent;
  - (14b) `extras/` includes `Future.request.yaml` with `$kind: carrier-pigeon-request` → an
    `unsupported_request_type` warning, not an error;
  - (16) `apiSecret` (`secret: true`) in the vault, with `secret: true` and no value in the collection file;
  - the zip path for v3 folders is covered by `workspace.zip` (T043) and `large-1000-v3.zip` (T053).

  Prove red.
- [ ] T031 [P] [US2] In `test/integration/features/import/format_detection.feature`, move the `postman v3 folder`
  and `postman v3 file` rows from the `@refused` outline to the `@detect` outline as
  `| postman/v3/native-git | postman_collection | 3 |` and `| postman/v3/native-git/Get users.request.yaml | postman_request | 3 |`. Add
  `| postman/v3/rich.json | postman_collection | 3 |` and `| postman/v3/rich | postman_collection | 3 |`. Prove the
  moved rows red.

### Implementation

- [ ] T032 [US2] In `internal/integration/persistence/import_source.go`:
  - `isImportDocument` allows the `.resources` segment (R12) and still skips every other dot segment;
  - `readImportFile`, for a name ending with `.request.yaml`, also reads
    `<dir>/.resources/<stem>.resources/examples/*.example.yaml` into `Documents`, with paths relative to `<dir>`;
  - YAML files get `probeDocument`-style probing into `model.ImportProbe` through `yaml.Unmarshal`. A YAML parse
    error leaves the probe as today (`{Name}`).

  Update `import_source` unit tests.
- [ ] T033 [US2] Detection (R4.2, R4.3, R4.5):
  - in `internal/domain/entity/import_probe.go`, replace `IsPostmanV3File`/`HasPostmanV3Entry` with
    `IsPostmanV3Request()` and `IsPostmanV3CollectionFolder()`;
  - in `internal/application/usecase/detect_import_format.go`:
    - the `IsPostmanV3File` branch returns `importkind.PostmanRequest, "3"`;
    - `detectFolder`'s v3 branch returns `importkind.PostmanCollection, "3"`;
    - a YAML probe with `PostmanKind == "collection"` returns `PostmanCollection, "3"`;
    - delete `postmanV3Msg`;
  - update `import_detection_test.go` rows `:54-55`.
- [ ] T034 [US2] Create `internal/integration/webservice/postmanv3/dto/definition.go` (type `Definition`) and
  `tree.go`. `NewTree(documents []entity.ImportDocument, root string)` assembles a folder into a `Collection`:
  - definitions name folders (R11);
  - `*.request.yaml` become requests (`IMP-11400` on decode error or unknown `$kind`);
  - examples are attached from the request's `examples` path, or else the `.resources/<stem>.resources/examples/`
    convention;
  - siblings are ordered per R10: `order` ascending, missing last, ties by `strings.ToLower(name)` then byte order.

  Unit-test ordering and naming in `tree_test.go`.
- [ ] T035 [US2] In `internal/integration/webservice/postman_v3.go`, handle a `PostmanCollection` folder input
  through `NewTree` → `ToPostmanV21` → the shared v2.1 step. Handle `PostmanRequest`:
  - build a one-request tree;
  - convert it through the shared v2.1 step wrapped in a synthetic collection;
  - return `PostmanV3Source.Request` with the request node (the first child) and its warnings.
- [ ] T036 [US2] Turn the US2 scenarios and the moved format-detection rows green. Run `make bdd TAG=@xen_009`,
  `make bdd TAG=@format_detection` and `make bdd TAG=@import`.

**Checkpoint**: Collection folders and request files import.

---

## Phase 5: User Story 3 — v3 environments and globals (P1)

**Goal**: v3 environment and globals YAML and JSON import, with secrets in the vault.

**Independent test**: `env/staging.environment.yaml` imports. No fixture secret appears in any workspace file or log.
Reveal returns the values.

### Tests (write first, prove red)

- [ ] T037 [US3] Create `test/integration/features/import/postman_v3_environment_globals.feature`
  (`@xen_009 @import @postman_v3_environment_globals`) with plan BDD Coverage scenarios (1)–(5):
  - (1) `the log output is captured`, then `the log output should not contain "s3cr3t-fixture-env"`, and
    `no file under the workspace folder should contain "s3cr3t-fixture-env"` / `"s3cr3t-fixture-token"`;
  - `VaultService.Reveal` for `apiKey` and `token`;
  - the `linked` row `secret_state` is `missing`, with a `linked_secret` warning whose message is
    `linked is linked to a Postman vault secret; enter its value in Xenon`, and the `postman.source.provider`
    passthrough is kept;
  - the `cli-environment.json` secret `secret_state` is `missing`, and the environment file never contains
    `********`;
  - globals YAML, `cloud-globals.json` and `local-globals.json` detect as `postman_globals`/`3` and merge with clash
    rules;
  - `noname.environment.yaml` → environment named `noname`.

  Prove red.
- [ ] T038 [P] [US3] Add detect rows to `test/integration/features/import/format_detection.feature`:
  - `postman/v3/env/staging.environment.yaml` → `postman_environment 3`;
  - `postman/v3/env/cli-environment.json` → `postman_environment 3`;
  - `postman/v3/env/workspace.globals.yaml`, `cloud-globals.json` and `local-globals.json` → `postman_globals 3`.

  Prove red.

### Implementation

- [ ] T039 [US3] Detection (R5):
  - add to `internal/domain/entity/import_probe.go` `IsPostmanV3Variables()` and `IsPostmanV3Globals()` (file-name
    suffixes `.globals.yaml`/`.globals.yml`/`.globals.json`, `HasWorkspaceField`, or `DocumentID` ending with
    `.globals.yaml`);
  - in `internal/application/usecase/detect_import_format.go`, JSON (`detectJson`, after the `_postman_variable_scope`
    checks) and YAML probes return `PostmanEnvironment`/`PostmanGlobals` with version `3`;
  - add unit rows.
- [ ] T040 [US3] Create `internal/integration/webservice/postmanv3/dto/environment.go` (`Environment`, `Value`).
  `ToPostmanV21(scope, fileStem, conversion)` builds `{name, values, _postman_variable_scope}`:
  - the name falls back to the file stem without `.environment.yaml`;
  - `secret: true` or `type: secret` → `type: "secret"`;
  - a `********` value on a secret → `""`;
  - `source` present → `value: ""`, `type: "secret"`, `source` copied, plus a `linked_secret` warning (R8 message,
    path = environment name / key);
  - other keys are copied.
- [ ] T041 [US3] In `internal/integration/webservice/postman_v3.go`, handle `PostmanEnvironment`/`PostmanGlobals`
  v3: decode JSON or YAML (on error, `IMP-11400` through `DecodeError`), then `ToPostmanV21`, then the existing `dto.Environment` → `ToEntity`. Fill
  `Environments` or `Globals` and add the conversion warnings to `Warnings`.
- [ ] T042 [US3] Turn the US3 scenarios and detect rows green. Run `make bdd TAG=@xen_009` and
  `make bdd TAG=@postman_environment_globals`.

**Checkpoint**: Environments and globals import, and secrets are safe.

---

## Phase 6: User Story 4 — whole workspace folder (P2)

**Goal**: A `postman workspace pull` folder imports every collection, environment and the globals in one go.

**Independent test**: `v3/workspace/` detects five items and two skipped parts, and imports into two workspaces.

### Tests (write first, prove red)

- [ ] T043 [US4] Create `test/integration/features/import/postman_v3_workspace.feature`
  (`@xen_009 @import @postman_v3_workspace`) with plan BDD Coverage scenarios (1)–(6):
  - (1) `items` are Billing, Users, prod, staging, Globals with kinds `collection, collection, environment, environment, globals`;
  - (3) `warnings` contains `skipped_part` for `flows` and `specs` with the exact R8 message;
  - (4) picking `v3/workspace` succeeds even though `tools/latin1.json` is not UTF-8;
  - (5) picking `v3/workspace/postman/collections` lists only the two collections;
  - (6) `workspace.zip` gives the same preview;
  - (7) a temp copy of the workspace with a broken environment file (written with `the file … exists with
    content`) is refused at detect with `IMP-11400` naming that file, and nothing is written;
  - (8) a temp folder holding only `postman/specs/a.yaml` → `IMP-00415` "… is not a Xenon, Postman, Bruno or cURL
    export";
  - (2) import with per-item `target_local_id` across workspaces "Acme" and "Beta", each collection, environment and
    globals file landing in its target.

  Prove red.
- [ ] T044 [P] [US4] Add detect rows to `format_detection.feature`: `postman/v3/workspace`,
  `postman/v3/workspace/postman`, `postman/v3/workspace/postman/collections` and `postman/v3/workspace.zip` →
  `postman_workspace 3`. Prove red.
- [ ] T045 [P] [US4] Vitest:
  - in `frontend/src/features/import-export/lib/labels.test.ts`, `sourceLabel('postman_workspace', '3')` is
    "Postman workspace v3" and `sourceLabel('postman_request', '3')` is "Postman request v3";
  - add a case to `frontend/src/features/import-export/components/ImportDialog.test.tsx`, in the existing
    `describe('ImportDialog')` with its mocked `platform/`: a detected `postman_workspace` preview renders the
    per-item workspace table and the "Not Imported" list.

  Prove red (`npm --prefix frontend run test`).

### Implementation

- [ ] T046 [US4] In `internal/integration/persistence/import_source.go` `readImportFolder`, if
  `<root>/postman/collections`, `<root>/postman/environments` or `<root>/postman/globals` is a directory, walk
  `<root>/postman` instead of `<root>` (R4 re-rooting). The probe `Name` stays the picked folder's name. Add a unit
  test.
- [ ] T047 [US4] Detection (R4.1):
  - add `IsPostmanV3Workspace()` to `internal/domain/entity/import_probe.go`, covering top-level
    `collections/<c>/` with v3 entries, `environments/*.environment.yaml`, `globals/*.globals.yaml`, or a folder
    named `collections` whose subfolders hold v3 entries;
  - in `internal/application/usecase/detect_import_format.go` `detectFolder`, check it before the v3 collection rule
    and return `importkind.PostmanWorkspace, "3"`;
  - add unit rows.
- [ ] T048 [US4] Create `internal/integration/webservice/postmanv3/dto/workspace.go`. `NewWorkspace(documents,
  entries, probeName)` splits:
  - collections: each directory directly under `collections/`, or under the root when the probe is named
    `collections`, becomes `NewTree`;
  - environments: `environments/*.environment.yaml|yml|json`;
  - globals: `globals/*.globals.yaml|yml|json`;
  - skipped parts: other top-level parts with file counts, sorted.

  Collections are sorted by folder name and environments by file name (contracts/import-service.md). In
  `internal/integration/webservice/postman_v3.go`, handle `PostmanWorkspace` by converting each part with the US1–US3
  helpers into `PostmanV3Source`.
- [ ] T049 [US4] Frontend:
  - in `frontend/src/features/import-export/lib/labels.ts`, add `postman_request: 'Postman request'` and
    `postman_workspace: 'Postman workspace'`;
  - in `frontend/src/features/import-export/components/ImportPreviewStep.tsx`, replace `DUMP_KIND` with
    `ASSIGNMENT_KINDS = new Set(['postman_dump', 'postman_workspace'])` in both conditions.
- [ ] T050 [US4] Turn US4 green. Run `make bdd TAG=@postman_v3_workspace`, `make bdd TAG=@format_detection`,
  `make bdd TAG=@postman_dump` and the frontend lint, typecheck and Vitest.

**Checkpoint**: Whole workspaces import.

---

## Phase 7: User Story 5 — MCP `import` tool parity (P2)

**Goal**: The MCP `import` tool gives the dialog's results for v3 sources.

**Independent test**: `detect` and `run` on the workspace fixture through a real MCP client.

- [ ] T051 [US5] Add a scenario to `test/integration/features/mcp/other_tools.feature` (tags
  `@xen_009 @mcp_import_v3`):
  - `import` `detect` on `{{resources}}/import/postman/v3/workspace` returns kind `postman_workspace`, version `3`,
    5 items and the two `skipped_part` warnings;
  - `run` imports every item into the active workspace;
  - the tool description contains "v3".

  Follow the existing import scenario at `:15-33` for step usage. Prove red.
- [ ] T052 [US5] Update the `import` tool description in `internal/integration/entrypoint/controller/mcp_import.go`
  to the text in contracts/import-service.md. Turn T051 green.

---

## Phase 8: Polish & Cross-Cutting

- [ ] T053 [P] Add the v3 large-import scenario to `test/integration/features/import/large_collection.feature`:
  `postman/v3/large-1000-v3.zip` with Detect < 2000 ms, `items.0.counts.requests` 1000, Import < 8000 ms and
  `import:progress` at most 250 ms apart. Prove red, then green. It needs no new code; it is green once US2 is done.
- [ ] T054 Run the open-items sweep on code comments (`grep -rn "Postman v3 YAML is not supported" internal test frontend`).
  It must return nothing. The removed message must not survive in any test or fixture README.
- [ ] T055 Run `make sort-imports`, then `make validate` (Go unit + BDD, frontend lint, typecheck and Vitest). Every
  pre-existing scenario must be green (SC-005).
- [ ] T056 Walk `specs/XEN-009-postman-v3-import/quickstart.md` steps 1–6 in the desktop app.

---

## Dependencies & Execution Order

- Phase 1 (fixtures, PRD) and Phase 2 (plumbing) come before every story. T001 → T002 → T003 (generated from each
  other). T004–T008 and T010 are parallel. T016 depends on T013 and T017; do T017 first. T020 depends on T019. T021 depends on T014,
  T018 and T020.
- US1 (Phase 3) is first: it builds the request, body, auth and example DTOs every later story reuses.
- US2 depends on US1's DTOs (T027, T028).
- US3 depends only on Phase 2 and can run in parallel with US2.
- US4 depends on US2 (`NewTree`) and US3 (environment DTO).
- US5 depends on US4's fixture and detection.
- Polish comes last.

## Parallel opportunities

- Phase 1: T004, T005, T006, T007, T008 and T010.
- Phase 2: T011–T015.
- US1: T025 alongside T024.
- US2: T031 alongside T030.
- US3: T038 alongside T037. US3 as a whole alongside US2.
- US4: T044 and T045 alongside T043.

## Implementation Strategy

1. **MVP = Phase 1 + Phase 2 + US1**: v3 JSON from `postman collection get --json` imports with v2.1 fidelity. That
   alone unblocks AI clients using the Postman CLI.
2. Then US2 (git folders) and US3 (environments and globals), which are independent of each other.
3. Then US4 (whole workspace), US5 (MCP proof) and Polish.
4. Each story ends with its tag green **and** `make bdd TAG=@import` green, so no story regresses existing imports.
