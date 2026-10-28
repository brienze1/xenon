# Claude Code Rules for Xenon

Xenon is a local-first desktop API client (a Postman replacement) built with **Wails v3**:
a Go backend that does all the real work and a React + TypeScript + Vite UI that only
displays and edits. Repo: `github.com/brienze1/xenon`. The product source of truth is the
PRD at `docs/prd/xenon-mvp-prd.md`.

**This is a single repo for everything** — backend, frontend, BDD tests and Spec Kit specs
all live here. There is no separate e2e, QA or spec repo.

## Repository layout

```
xenon/
├── cmd/                     ← Go entry points (desktop app; later a `xenon` CLI)
├── internal/                ← Go backend — Clean Architecture, see below
│   ├── domain/
│   ├── application/
│   ├── integration/
│   └── infra/
├── pkg/                     ← Reusable Go helpers (see the pkg-*.md skills)
├── frontend/                ← React + TS + Vite UI (Wails convention) — see frontend skills
├── build/                   ← Wails build assets (icons, platform config)
├── test/integration/        ← Godog BDD suite (features/, steps/, mock/, resources/)
├── specs/<ID>-<slug>/       ← Spec Kit feature specs (spec, plan, tasks, checklists)
├── docs/                    ← PRD, brand assets, design notes
├── .specify/                ← Spec Kit scripts/templates/constitution
└── .claude/                 ← agents, commands, skills
```

## Backend — Go, Clean Architecture (strict)

Custom 4-layer Clean Architecture (**not** Uncle Bob's standard layout) with DDD:

- **Domain** — entities, enums, business rules. Depends on nothing.
- **Application** — use cases, services, adapter interfaces. Depends on domain only.
- **Integration** — entrypoints, DTOs, models, persistence, webservices, publishers.
- **Infrastructure** — app startup, properties, dependency injection, wiring.

Dependencies point inward, always. The canonical rules live in `.claude/skills/` — start at
`.claude/skills/architecture.md` ("where does this go") and follow the per-concern skill for
whatever you are writing (`entities.md`, `usecases.md`, `services.md`, `controllers.md`,
`dtos.md`, `models.md`, `persistence.md`, `webservices.md`, `publishers.md`, `properties.md`,
`dependency-injection.md`, `naming-conventions.md`, `comments.md`, …).

**Every backend change follows these skills strictly. The skills are law.** When existing code
and a skill disagree, the skill is right and the code is wrong. No exemptions for "existing
pattern", "project convention", or "we'll fix it later".

How the PRD's tech rules map onto the layers:

- **Wails bindings play the REST router's role.** Each Go service React calls (workspaces,
  tree, requests, history, UI state, …) is a thin binding struct in
  `internal/infra/server/binding/` that only adapts (ctx, DTO in → controller → DTO out). The
  controllers live in `internal/integration/entrypoint/controller/` and keep the full controller
  skill pattern (validation, `ToEntity`, application adapter, error handler, panic recovery).
  Only `internal/infra/` and `cmd/` may import Wails.
- **"The Go core has no Wails imports"** is satisfied by the layering itself: domain,
  application and the non-entrypoint integration packages never import Wails, so a future
  `cmd/cli` reuses them by wiring a different entrypoint.
- Disk (YAML), the OS keychain, the vault, git (go-git), the HTTP engine and goja are all
  outer boundaries reached through adapter interfaces — persistence / webservice
  implementations in `internal/integration/`, never called directly from use cases.
- React never touches the disk or the network directly.

## Frontend — React + TypeScript (Wails)

`.claude/skills/frontend-architecture/SKILL.md` is the frontend's `architecture.md`: folder
structure (`app/` · `features/` · `platform/` · `shared/`), import boundaries, the Wails
bridge, Zustand state, styling tokens, hotkeys and testing. **Every frontend change follows it
strictly** — same weight as the backend skills. Alongside it, the Vercel skills:
`vercel-react-best-practices`, `vercel-composition-patterns`, `vercel-web-design-guidelines`
(vendored from `vercel-labs/agent-skills`). Next.js/server-only rules in them do not apply —
this is a client-only SPA in a desktop webview.

## BDD-First Development

Tests come before implementation, without exception:

1. Write the Gherkin scenarios that encode the expectation (`bdd-feature-files.md`).
2. Run them and **prove they are red**. A scenario that passes before the code exists is not
   testing the new behaviour.
3. Implement innermost layer outward until green.
4. Run the full suite — new scenarios *and* every pre-existing one — and keep it green.

The harness lives in `test/integration/` — `make bdd`, or `make bdd TAG=@your_tag`. Only the
outermost boundary is faked (filesystem/keychain/HTTP targets/git remotes via the mocks in
`test/integration/mock/`); persistence, services, use cases and entrypoints are never mocked.

**Step definitions are a shared, generic catalog.** Write new features against the steps that
already exist in `test/integration/steps/`; only extend the catalog when a genuinely new kind of
assertion is needed, and keep the new step generic. See `bdd-testing.md`,
`bdd-feature-files.md`, `bdd-step-definitions.md`, `testing-setup.md`, and the `mock-*.md` skills.

## Agents

- Backend implementation: `@agent-golang-clean-arch-specialist` (user-level agent).
- Pre-merge review: `@agent-golang-clean-arch-reviewer` (user-level agent).
- Refactor pipeline (used by `/refactor-diff`): `orchestrator`, `scanner`, `executor`,
  `tester`, `reviewer`, `skill-expert`, `infer` — in `.claude/agents/`.

## Commands — the task flow

Tasks are **GitHub issues** on `brienze1/xenon` (there is no Linear), read and updated with `gh` per
`.claude/skills/github-issues.md`. Each task has an id `XEN-NNN` in its issue title. All four commands run on
**one branch per task** (the lowercased id, e.g. `xen-008`):

```
/research-task → /spec-task → /create-tests → /implement-task
  (issue desc)     (specs/)     (red .feature)   (green code → push → PR → merge on green CI)
```

- `/research-task <issue>` — verify the issue against the code, resolve questions, rewrite the
  issue description. No code.
- `/spec-task <issue>` — full Spec Kit chain (specify → clarify → plan → tasks → analyze →
  checklist) into `specs/<XEN-ID>-<slug>/`, zero open questions, commit + push the spec.
- `/create-tests <issue>` — red BDD features in `test/integration/features/` from the spec,
  commit + push to the task branch.
- `/implement-task <issue | description>` — branch fresh off `origin/main` (reusing the task
  branch), red → green, `/refactor-diff` until clean, push, then opens a PR into `main` and merges it
  once GitHub Actions pass (always; never merges red).
- `/refactor-diff` — PR-diff-scoped refactor; touches only files changed vs `main` plus
  uncommitted work. Autonomous, never commits.

Spec Kit skills (`speckit-*`) are also available directly.

## Conventions

- Never commit `REFACTORING_REPORT.md` or other refactor artifacts.
- Work in a dedicated worktree branched from a **freshly fetched** `origin/main`.
- Issue comments follow `.claude/skills/github-issues.md`.
- Build/validation targets follow `.claude/skills/makefile.md` — `make validate` is the final
  quality gate, `make sort-imports` after each batch of Go changes.
- Secrets never land in workspace files, logs, fixtures or commits (PRD: Environments and secrets).

## CI and flaky tests

- **Test locally before waiting on CI.** Run the affected tests (and `make validate` before pushing) on this
  machine first; it is far faster than a GitHub Actions round trip. CI confirms, it is not the first test run.
- **A failing CI check is ours to fix, whatever its scope.** If a check fails on a branch you are working on, fix the
  cause in that PR (or a dedicated fix branch) even when the failure has nothing to do with the task. Never merge red,
  never skip or disable a check, never wave a failure through as "unrelated".
- **A flaky test is worse than a failing one** — it trains everyone to ignore red. Give flakes double focus: find
  the real cause (timing assumptions, debounce windows, shared state, ordering, slow runners), reproduce it locally
  (e.g. inject the delay, loop the test with `-count=N`), fix it so the failure mode is impossible, and prove the
  fix by showing the reproduction fails before and passes after. Re-running a job is never the fix; at most a
  stopgap while the fix is made.
- Performance budgets (`the last call should have taken less than N ms`) are product requirements: fix a slow path
  by making the code faster, not by raising the budget.

## Context7 usage

Always use Context7 when you need code generation, setup or configuration steps, or
library/API documentation (Wails v3, goja, go-git, age, go-keyring, CodeMirror 6, dnd-kit,
Zustand, …). Resolve library ids and fetch docs automatically, without being asked.
