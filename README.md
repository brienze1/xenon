# Xenon

Xenon is a local-first desktop API client: a Postman replacement with no account, no cloud
and no lock-in. Every workspace is a plain folder of YAML files (one file per request) that you
can commit to your own git repository; Xenon only adds what it needs to remember locally (open
tabs, layout, send history) in its app-data folder on your machine.

It is built with [Wails v3](https://v3alpha.wails.io/): a Go backend does all the real work
(sending requests, reading and writing files, history) and a React + TypeScript UI only displays
and edits. The product source of truth is the PRD in
[`docs/prd/xenon-mvp-prd.md`](docs/prd/xenon-mvp-prd.md).

## Prerequisites

| Tool | Version |
| --- | --- |
| Go | 1.26 or newer (the toolchain pinned in `go.mod` is fetched automatically) |
| Node.js | current LTS, with npm |
| make | any recent version (on Windows use Git Bash, MSYS2 or WSL) |

Platform prerequisites for the Wails webview:

- **macOS**: Xcode Command Line Tools (`xcode-select --install`).
- **Windows 10/11**: the WebView2 runtime (preinstalled on current Windows).
- **Linux** (Ubuntu 24.04 or similar): `sudo apt install build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev`.

## Getting started

```bash
make setup   # installs the pinned wails3 CLI and Go tools, then npm ci in frontend/
make dev     # runs the app with hot reload (Go + Vite)
```

Other useful targets:

| Command | What it does |
| --- | --- |
| `make build` | builds the desktop app for the current OS into `bin/` |
| `make bindings` | regenerates the TypeScript bindings in `frontend/bindings/` (git-ignored) |
| `make bdd` / `make bdd TAG=@history` | runs the Godog BDD suite in `test/integration/`, optionally by tag |
| `make tests` | Go unit + BDD tests with coverage |
| `make frontend-check` | frontend lint, typecheck and Vitest |
| `make validate` | the full quality gate (imports, alignment, vet, dead code, vulnerabilities, tests, coverage, mutation testing, frontend checks) |

## Where your data lives

**Workspace data** is the folder you pick when you create or open a workspace: `workspace.yaml`,
`collections/` (one folder per collection and folder, one YAML file per request) and
`environments/` (one YAML file per environment). Variables are stored as rows with their
**initial values** only: globals in `workspace.yaml`, collection variables in `collection.yaml`,
environment variables in their environment file. A secret row keeps only its name, `secret: true`,
enabled and description — never a value. The format is described in
[`specs/XEN-001-shell-and-storage/contracts/workspace-format.md`](specs/XEN-001-shell-and-storage/contracts/workspace-format.md)
and [`specs/XEN-002-environments-and-secrets/contracts/format-additions.md`](specs/XEN-002-environments-and-secrets/contracts/format-additions.md).
Nothing else is ever written into a workspace folder, so it is safe to commit.

**Local data** stays on this machine, in Xenon's app-data folder:

| OS | Folder |
| --- | --- |
| macOS | `~/Library/Application Support/Xenon/` |
| Windows | `%AppData%\Xenon\` |
| Linux | `$XDG_CONFIG_HOME/xenon/` (default `~/.config/xenon/`) |

```
Xenon/
├── app.json                   ← registered workspaces, the active one, window size and position,
│                                vault mode (keychain | passphrase) and the passphrase check
├── logs/xenon.log             ← rotated log (never contains header values, auth, bodies or secrets)
└── workspaces/<localId>/
    ├── state.json             ← open tabs with unsaved drafts, sidebar state and layout
    ├── history.jsonl          ← the 1,000 most recent sends of that workspace (unresolved {{variables}})
    ├── variables.json         ← the active environment and the current values of plain variables
    ├── vault.age              ← secret values, encrypted with age to this workspace's vault key
    └── vault-key.age          ← passphrase mode only: the vault key, encrypted with your passphrase
```

On macOS and Linux these folders are owner-only (`0700`) and the files `0600`. Set
`XENON_DATA_DIR` to use another folder (the BDD suite does this to stay isolated).

**Secrets.** Each workspace has its own vault key. Normally it lives in the OS keychain (macOS
Keychain, Windows Credential Manager, Linux Secret Service) under the service
`dev.brienze.xenon.vault` with the workspace's local id as the account, so no password is
needed. When the keychain is unavailable Xenon asks once for a passphrase (stored nowhere; asked
again after each start to unlock) and keeps the key in `vault-key.age` instead. Removing a
workspace deletes its `variables.json`, vault files and keychain entry. A lost key or forgotten
passphrase cannot be recovered: "Reset vault" starts an empty vault and the secrets show as
missing until you type them again.

## Repository layout

```
cmd/xenon/         desktop entry point
internal/          Go backend, Clean Architecture: domain → application → integration → infra
pkg/               standalone Go helpers
frontend/          React + TypeScript + Vite UI (app/ · features/ · platform/ · shared/)
test/integration/  Godog BDD suite (features/, steps/, mock/, resources/)
specs/             Spec Kit feature specs
docs/              PRD and brand assets
```

Contributors: read [`CLAUDE.md`](CLAUDE.md) and the skills in `.claude/skills/` before changing
code; they define where every piece of code goes and the BDD-first workflow.
