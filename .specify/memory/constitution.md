<!--
Sync Impact Report
- Version change: (template) → 1.0.0
- Principles defined: I. Local-First, Files Are the Data · II. Secrets Never Leak ·
  III. Go Owns All Logic · IV. Architecture Skills Are Law · V. BDD-First ·
  VI. Postman Parity, Nothing More · VII. Cross-Platform From Day One
- Added sections: Technology & Design Constraints, Development Workflow, Governance
- Removed sections: none
- Templates:
  ✅ .specify/templates/plan-template.md — Constitution Check gates filled in
  ✅ .specify/templates/tasks-template.md — tests are mandatory (BDD-first), not optional
  ✅ .specify/templates/spec-template.md — no change needed
  ✅ CLAUDE.md — consistent
- Deferred TODOs: none
-->

# Xenon Constitution

## Core Principles

### I. Local-First, Files Are the Data

- Xenon MUST work with no account, no Xenon cloud and no network service of its own.
- Every workspace MUST be a plain folder on disk; collections, folders, requests and
  environments are YAML files (one file per request) mirroring the sidebar. The files are the
  data — there is no hidden database that can disagree with them.
- Export and import MUST be free at every level (workspace, collection, folder, request,
  environment), as Xenon YAML and as Postman JSON.
- A workspace is a hard boundary: nothing (requests, tabs, environments, cookies, history)
  from one workspace may appear in or be reachable from another.

Rationale: the product exists because the alternatives lock data in a cloud or behind paywalls.

### II. Secrets Never Leak

- Secret values MUST NOT be written to workspace files, plain exports, logs, the console,
  code snippets, test fixtures or git commits. Files store the variable name and
  `secret: true` only.
- Secret values live in one age-encrypted vault file per workspace in Xenon's app-data
  folder; the vault key lives in the OS keychain.
- Masking (`•••••`) is the default everywhere a value is displayed or copied; revealing or
  including a secret is always an explicit user action.
- Exports MAY include secrets only when the user opts in and supplies a passphrase; they are
  then encrypted.

Rationale: workspaces are meant to be committed to company repos; one leaked token is a breach.

### III. Go Owns All Logic

- Sending requests, running scripts, file I/O, secrets, cookies and git MUST be implemented
  in Go. The React UI only displays and edits state and MUST NOT touch the disk or the
  network directly.
- The UI talks to Go only through Wails-generated bindings and events, wrapped by
  `frontend/src/platform/`.
- Domain, application and non-entrypoint integration packages MUST NOT import Wails, so a
  future `xenon` CLI reuses them by wiring a different entrypoint.
- Scripts run in goja inside Go, sandboxed (no file system, no shell), with a time limit —
  never in the webview.

Rationale: one engine, testable without a UI, identical behaviour in the app and a future CLI.

### IV. Architecture Skills Are Law (NON-NEGOTIABLE)

- Backend code MUST follow the 4-layer Clean Architecture skills in `.claude/skills/`
  (`architecture.md` and the per-concern skills). Dependencies point inward only.
- Frontend code MUST follow `.claude/skills/frontend-architecture/` (folder structure and
  import boundaries) and the Vercel skills (`vercel-react-best-practices`,
  `vercel-composition-patterns`, `vercel-web-design-guidelines`).
- When code and a skill disagree, the code is wrong. No exemptions for "existing pattern" or
  "later".

Rationale: one predictable place for everything keeps a large, mostly agent-written codebase
navigable.

### V. BDD-First (NON-NEGOTIABLE)

- Every behaviour is first written as Godog scenarios in `test/integration/features/` and
  proven red before implementation, then implemented to green.
- Only the outermost boundary is faked (file system roots, OS keychain, remote HTTP targets,
  git remotes, clock); persistence, services, use cases and entrypoints are never mocked.
- The full suite (Go unit + BDD, frontend lint + typecheck + Vitest) MUST be green before a
  branch is pushed. Frontend tests mock `platform/`, never the generated bindings.

Rationale: Postman parity is a long list of exact behaviours; only executable specs hold it.

### VI. Postman Parity, Nothing More

- Layout, hotkeys, tab behaviour, variable scopes and precedence, `pm.*` script semantics and
  script execution order MUST match Postman so muscle memory and existing collections carry
  over unchanged.
- Postman → Xenon → Postman MUST round-trip without losing requests, scripts, auth or
  variables.
- Anything not in the MVP PRD (`docs/prd/xenon-mvp-prd.md`) is out of scope until the PRD
  changes. Visual identity is Xenon's own (PRD palette, dark theme only).

Rationale: the MVP's definition of done is "Luis uninstalls Postman for a full working week".

### VII. Cross-Platform From Day One

- Xenon MUST build, run and pass its test suite on macOS, Windows and Linux.
- `Cmd` shortcuts map to `Ctrl` on Windows and Linux; keychain access uses each OS's native
  store (Keychain, Credential Manager, Secret Service).
- Paths, line endings and file names MUST be handled portably (no hard-coded separators,
  safe file names for request/collection names on every OS).

Rationale: decided on 2026-10-09; retrofitting portability later costs more than building it in.

## Technology & Design Constraints

| Part | Choice |
| --- | --- |
| Desktop shell | Wails v3 (pinned latest beta; move to 3.0 when it ships) |
| UI | React + TypeScript + Vite, Zustand, CodeMirror 6, dnd-kit, CSS Modules + design tokens |
| HTTP engine | Go `net/http` + `httptrace`, cookie jar per workspace |
| Scripts | goja with a Go-implemented `pm` object and bundled Chai for `pm.expect` |
| Storage | YAML via `yaml.v3`, one file per request |
| Secrets | age-encrypted vault per workspace, key in OS keychain via `go-keyring` |
| Git sync | go-git (pure Go) |
| Import/export | Own Go converters for Postman collection v2.1 / environment JSON, and Bruno |

- Colours, glows and fonts come only from the PRD palette tokens.
- New dependencies beyond this table require justification in the feature's plan.

## Development Workflow

- Single repo: code, BDD suite, frontend and Spec Kit specs (`specs/<ID>-<slug>/`) live
  together.
- Task flow, one branch per task: `/spec-task` → `/create-tests` → `/implement-task`.
- Branches start from a freshly fetched `origin/main`.
- Quality gates: `make validate` and the frontend `lint`, `typecheck` and `test` scripts;
  `/refactor-diff` until clean before pushing.
- `REFACTORING_REPORT.md` and other refactor artifacts are never committed.

## Governance

- This constitution supersedes other practices and guidance; `CLAUDE.md` and the skills
  implement it.
- Every `plan.md` MUST pass the Constitution Check before research and again after design;
  any violation is recorded in its Complexity Tracking table with justification.
- Amendments are made by pull request, update the Sync Impact Report, and bump the version:
  MAJOR for removed or redefined principles, MINOR for new principles or sections, PATCH for
  clarifications.

**Version**: 1.0.0 | **Ratified**: 2026-10-09 | **Last Amended**: 2026-10-09
