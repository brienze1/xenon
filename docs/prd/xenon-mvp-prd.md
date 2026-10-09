# Xenon — MVP PRD

Oct 9, 2026 · @Luis Brienze

## Overview

Xenon is a local-first desktop API client that works like Postman, with no account, no cloud and no paywalls. It replaces Postman and Bruno for one user who works across 4 companies.

**The problem.** Postman puts export and sharing behind a paywall and pushes everything to its cloud. Bruno is file-based and exports well, but limits the free version to 2 open workspaces. Neither keeps 4 companies cleanly separated for free.

**Who it's for.** Luis first: one person, several companies, daily API work. Open-sourcing it can come later and is not a goal of this MVP.

**What "usable" means.** Luis can uninstall Postman and do a full working week in Xenon without going back.

## Goals and non-goals

The MVP is done when everything Luis does in Postman daily works in Xenon, and nothing beyond that is built.

**Goals**

- Real workspace separation: one per company, each with its own collections and environments.
- Postman-level request building, scripting and response inspection.
- Everything stored as plain files on disk, with free export and import (YAML and Postman).
- Secrets that stay out of exported files and out of git.
- Postman muscle memory works: same layout, same hotkeys, same tab behaviour.

**Non-goals for the MVP**

- Accounts, a Xenon cloud, live collaboration or teams. Sharing goes through Git sync and export.
- Mock servers, monitors, API docs generation, API design (OpenAPI editing).
- Protocols other than HTTP: GraphQL gets a body type only; gRPC, WebSocket and SSE come later.
- Collection runner and CLI (useful, but not needed to replace daily use).
- Visual polish beyond clean and usable: colours and fonts are Xenon's own.

## Screen layout

Xenon keeps Postman's screen placement so nothing has to be relearned; only colours and fonts differ.

&#91;embedded content: Xenon main window · region placement\]

- The sidebar, right panel and console can each be hidden, and their sizes are remembered.
- The split between request and response can be dragged, and switched to side-by-side.

## Visual identity

Xenon looks like a dark club wall at night: near-black, shadowy surfaces with neon green, cyan and magenta taken from the app icon, and a graffiti edge in the details. Neon is for accents and state, never for long text, so the app stays readable for a full working day.

**Palette**

| Token | Hex | Used for |
| --- | --- | --- |
| Background | #0A0A0C | Window, behind everything |
| Surface | #141418 | Sidebar, panels, editors |
| Surface raised | #1E1E24 | Tabs, menus, dialogs, hover |
| Border | #2A2A33 | Dividers, input outlines |
| Text | #E8E8F0 | Body and labels |
| Text muted | #8A8A99 | Secondary text, placeholders |
| Neon green | #3DFF4F | Primary accent: Send button, active tab, success, 2xx |
| Neon cyan | #3EC8FF | Info, links, focus rings, selected tree item |
| Neon magenta | #FF3DFF | Secondary accent: scripts, variables, highlights |
| Neon yellow | #FFE53D | Warnings, 3xx, unsaved-changes dot |
| Neon red | #FF3D5A | Errors, 4xx and 5xx, DELETE, prod environment |

**HTTP method colours**

GET green, POST magenta, PUT cyan, PATCH yellow, DELETE red, HEAD and OPTIONS muted grey. Used on the method badge in the sidebar, tabs and request bar.

**Neon and shadow rules**

- Glow (a soft blurred halo in the accent colour) only on things that are active or need attention: the Send button, the active tab underline, the focused input, the environment badge, a running request.
- Panels are separated by depth and soft shadows, not bright lines.
- Each environment can take a neon colour; prod defaults to red, so the switcher glows red when prod is active.
- Dark theme only in the MVP. A light theme is out of scope.

**Graffiti touches**

- A marker-style display font only for the Xenon wordmark, empty states and the welcome screen. Everyday UI text stays in a clean sans (Inter), and code, bodies and URLs in a monospace (JetBrains Mono).
- Empty states (no workspace, no requests, no response yet) get a spray-paint style illustration or tag in the neon palette, with a subtle noise texture behind.
- Everything else stays clean: no textures behind tables, editors or responses.

**App icon**

![Xenon app icon](../brand/xenon-icon.svg)

Files: `docs/brand/xenon-icon.svg`, `docs/brand/xenon-icon-1024.png` (Wails icon source).

Use Luis's final icon files as delivered (SVG and 1024px PNG); its neon green, cyan and magenta set the palette above. Wails builds the platform icon files from the 1024px PNG.

## Workspaces, collections and folders

A workspace is a hard boundary: switching workspace swaps the whole sidebar, the open tabs and the environment list. Nothing from one company shows up in another.

**Workspaces**

- Create, rename, delete and switch workspaces from a switcher at the top left.
- Each workspace is one folder on disk, which can be anywhere (for example inside that company's git repo).
- Each workspace remembers its own open tabs, selected environment and sidebar state.
- No limit on how many workspaces exist or are open.

**Collections and folders**

- A workspace holds collections; a collection holds folders and requests; folders can nest.
- Collections and folders have their own settings: auth, headers, variables, and pre-request and post-response scripts (see Scripts).
- Create, rename, duplicate and delete any item from the sidebar or its right-click menu.
- Sidebar search filters requests by name and URL within the current workspace.

**Drag and drop**

- Drag a request into another folder or collection to move it.
- Drag a request within a folder to change its order.
- Drag folders to reorder them or nest them in another folder.
- Collections can be reordered too.
- The order is saved to disk, so it survives a restart and an export.
- Moving between workspaces is not drag and drop; use export and import (or a later "Move to workspace" action).

## Requests and tabs

A request in Xenon has every option a Postman request has, laid out in the same sub-tabs, in the same order.

**Request bar**

- Method dropdown: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, plus a custom method typed in.
- URL field with `{{variable}}` highlighting: green when resolved, red when not, value shown on hover.
- Send button (with Cancel while a request is in flight) and Save.

**Request sub-tabs**

| Tab | What it holds |
| --- | --- |
| Params | Query params as a key/value table, synced both ways with the URL; path variables (`:id`) get their own table |
| Authorization | Inherit from parent, No auth, Bearer, Basic, API key (header or query), OAuth 2.0 (auth code with PKCE, client credentials, password, refresh), AWS Signature, Digest |
| Headers | Key/value table with enable checkboxes and descriptions; auto-generated headers shown greyed out and can be toggled |
| Body | none, form-data (text and files), x-www-form-urlencoded, raw (JSON, XML, text, HTML, JavaScript) with syntax highlighting, binary file, GraphQL (query + variables) |
| Scripts | Pre-request and Post-response editors (see Scripts) |
| Settings | Timeout, follow redirects, SSL verification on/off, encode URL automatically |

Every key/value table supports enable/disable per row, bulk edit as text, and `{{variables}}` in keys and values.

**Tabs**

- Clicking a request in the sidebar opens it in a tab. A single click opens a preview tab (italic title) that the next click replaces; editing or double-clicking pins it, as in Postman.
- New empty tab with the + button or Ctrl/Cmd+T; it can be sent without saving, then saved into a collection.
- Unsaved changes show a dot on the tab. Closing a changed tab asks to save, discard or cancel.
- Tabs can be dragged to reorder. Open tabs are restored per workspace on restart.

**Tab right-click menu**

- Duplicate tab
- Close tab
- Close other tabs
- Close all tabs
- Close tabs to the left / to the right
- Force close all (closes everything, discarding unsaved changes without asking)
- Reveal in sidebar

## Scripts

Scripts are a must-have, and they must run existing Postman scripts unchanged. Writing new scripts uses the same `pm.*` API, so nothing new has to be learned and imported collections keep working.

**Where scripts live**

- Collections, folders and requests each have a Pre-request script and a Post-response script.
- Both are JavaScript, edited in a code editor with highlighting and autocomplete for `pm.*`.

**Inheritance (optional)**

By default a request runs its parents' scripts around its own, outermost first. Each folder and request has an "Inherit parent scripts" toggle, on by default; turning it off skips everything above that level.

Order for one send, with inheritance on:

1. Collection pre-request
2. Folder pre-request (each nested folder, outer to inner)
3. Request pre-request
4. The request is sent
5. Collection post-response
6. Folder post-response (outer to inner)
7. Request post-response

This matches Postman's order, so imported collections behave the same.

**What scripts can do (the `pm` API subset for the MVP)**

- Read and set variables: `pm.environment`, `pm.collectionVariables`, `pm.globals`, `pm.variables` (resolved value across all scopes).
- Read and change the outgoing request: `pm.request.url`, `headers`, `body`.
- Read the response: `pm.response.code`, `status`, `headers`, `json()`, `text()`, `responseTime`.
- Tests: `pm.test("name", fn)` with `pm.expect` (Chai-style assertions); results shown in the response's Test Results tab.
- Chain calls: `pm.sendRequest(...)` from inside a script.
- `console.log/warn/error` output goes to the Xenon console.

**Safety and limits**

- Scripts run in a sandbox with no file system or shell access.
- A script error stops the send and shows the error and line in the console. A post-response error still shows the response.
- A script that runs longer than a time limit (for example 10 seconds) is stopped.

## Environments and secrets

Each workspace has its own environments, and secret values never touch the workspace files: they live in an encrypted vault on the machine, unlocked by a key kept in the OS keychain.

**Environments**

- Any number per workspace, named freely (local, dev, hml, prod...).
- Each variable row has: name, value, type (plain or secret), enabled, description.
- Variable scopes, from strongest to weakest: request-local (set in scripts) → environment → collection variables → workspace globals. Same precedence as Postman.
- Duplicate an environment to create another (for example copy hml into prod, then change the values).
- Optional per-environment "Confirm before send" flag, meant for prod.

**How secrets are stored**

- In the workspace files, a secret variable is saved with its name and `secret: true`, but no value. The files can go into git safely.
- The value is saved in a vault: one encrypted file per workspace, kept in Xenon's app-data folder, outside the workspace folder.
- The vault's key is stored in the OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service). Xenon reads it at startup, so secrets are available with no master password to type.
- One vault file per workspace instead of one keychain entry per secret avoids keychain clutter and repeated macOS permission prompts.

**How secrets behave in the app**

- Shown as dots, with an eye icon to reveal and a copy button.
- Resolved normally when a request is sent, so `{{api_key}}` just works.
- Masked as `•••••` in the console, in the variables panel, and in code snippets (with an option to include the real value when copying).
- Scripts can read them, and setting a secret from a script keeps it secret.
- On a new machine, or after import, a secret with no value is marked "missing" and requests using it warn before sending.

**Secrets on export**

- Excluded by default: the exported file has the names, not the values.
- Optional "Include secrets" asks for a passphrase and writes them encrypted into the export. Importing that file asks for the same passphrase.
- Postman environments marked as secret are imported as Xenon secrets.

## Response, console and side panel

The response area, console and right panel give the same information as Postman's, in the same places.

**Response viewer**

- Status line: status code and text (coloured by class), time, size, and a hover breakdown of timings (DNS, connect, TLS, first byte, download).
- Body tab with three views: Pretty (formatted and highlighted JSON, XML or HTML, with folding), Raw, and Preview (rendered HTML or image).
- Search inside the body, copy body, and save the response to a file.
- Cookies tab: name, value, domain, path, expiry, flags.
- Headers tab: response headers as a table.
- Test Results tab: pass/fail per `pm.test`, with a count on the tab.
- Large responses (over a few MB) open in Raw first to stay fast.

**Console**

- A bottom drawer, hidden by default, toggled by a button in the footer and a hotkey.
- Logs every request sent, including those from `pm.sendRequest`: method, URL, status, time, size.
- Each entry expands to show the resolved request (URL, headers, body) and the response (headers, body), exactly as sent and received after variables are replaced.
- Also shows script `console.log` output and errors, in order with the requests.
- Filter by type (requests, logs, errors), search, and clear. Secrets are masked.

**Cookie jar**

- Cookies are kept per workspace and sent automatically, like a browser.
- A Cookies dialog lists them by domain, to view, edit and delete.

**Right side panel**

A collapsible panel on the right of the request, with icon buttons to switch between:

- **Variables**: every `{{variable}}` used in the current request, with its resolved value and which scope it came from (environment, collection, global). Values can be edited in place.
- **Code**: the current request as a ready-to-copy snippet. MVP languages: cURL, fetch (JavaScript), Go net/http, Python requests. A copy button and an option to include or mask secrets.
- **Info**: request name, description (markdown), and where it lives in the collection.

## Import and export

Everything can go out and come back in, at any level, as Xenon YAML or as Postman JSON. All of it is free.

**Native storage**

Xenon's own format is YAML, one file per request, in folders that mirror the sidebar. A workspace folder looks roughly like this:

```
acme/
  workspace.yaml
  environments/
    dev.yaml
    prod.yaml
  collections/
    billing-api/
      collection.yaml
      invoices/
        folder.yaml
        create-invoice.yaml
```

Because the files are the data, the folder can be committed to git, zipped or synced, and an "export" of a workspace is just a copy of it.

**What can be exported, and how**

| What | Xenon YAML | Postman |
| --- | --- | --- |
| Whole workspace | One .zip of the workspace folder (collections + environments) | One Postman collection JSON per collection + one environment JSON per environment, in a .zip |
| One collection | .zip of the collection folder, or a single bundled .yaml | Postman collection v2.1 JSON |
| One folder | .zip or single bundled .yaml | Postman collection v2.1 JSON containing only that folder |
| One request | Its .yaml file | Postman collection v2.1 JSON with one request; or "Copy as cURL" |
| One environment | Its .yaml file | Postman environment JSON |

Export is on the right-click menu of every item, and in File → Export for the whole workspace. Secrets follow the rules in the Environments section.

**What can be imported**

- Xenon YAML: a workspace .zip or folder, a collection, a folder, or a single request.
- Postman collection v2.1 JSON (with its scripts, auth, variables and folder order).
- Postman environment JSON.
- Postman full data dump (Settings → Export data), mapped to one Xenon workspace per Postman workspace if possible.
- A cURL command pasted into the URL bar or the import dialog.

Import asks which workspace and which collection or folder to put things in. Name clashes ask to replace, keep both, or skip.

**Fidelity**

- Postman → Xenon → Postman should round-trip without losing requests, scripts, auth or variables.
- Xenon-only settings (like the "Inherit parent scripts" toggle) are dropped on Postman export, with a note in the export summary.

## Git sync

Each workspace can be linked to its own Git repository (GitHub or any other Git host), using any of your Git accounts, so each company's collections can live in that company's repo.

**Setup, per workspace**

- A Sync tab in the workspace settings: repository URL, branch, and an optional subfolder, so a workspace can live inside an existing repo (for example `api/xenon/` in the company's backend repo).
- Accounts are added once, in app settings (Git accounts), by GitHub sign-in in the browser, a personal access token, or the machine's SSH key. Each workspace then picks which account it syncs with. One account can serve every workspace, or each workspace can use a different one; any mix is allowed.
- The repo can belong to any organization or to a personal account with no organization, and workspaces can point at repos in different organizations, even under the same account. Xenon only needs the repo URL and an account with access to it.
- Credentials are kept in the OS keychain, never in the workspace files.
- A new workspace can be created by cloning a repo: paste the URL, pick the folder, done.
- A workspace with no repo set up works fully offline; sync is optional.

**What is synced**

- Collections, folders, requests, environments and workspace settings: the YAML files on disk.
- Never synced: secret values (they stay in the local vault), open tabs, window layout and other local UI state. Xenon writes a `.gitignore` for these.

**Daily use**

- A sync status in the footer for the current workspace: up to date, local changes, behind the remote, or conflict.
- A Sync panel listing changed requests and environments, with a commit message box and Commit and Push buttons, plus Pull.
- Optional settings per workspace: pull automatically when the workspace is opened, and push automatically after each save.
- Push and pull errors (no network, expired token, rejected push) show in the Sync panel with a clear next step.

**Conflicts**

- One file per request keeps conflicts rare: two people only clash when they edit the same request.
- On a conflict, Xenon lists the affected requests and, for each, offers Keep mine, Keep theirs or Keep both (saved as a copy). No line-by-line merge editor in the MVP.

**Not in the MVP**

- Branch switching and creating branches from inside Xenon.
- Pull requests, history view and blame.

## Hotkeys and environment switcher

Hotkeys copy Postman's, so muscle memory carries over; on Windows and Linux, Cmd becomes Ctrl.

**Hotkeys**

| Action | Shortcut |
| --- | --- |
| Send request | Cmd+Enter |
| Save | Cmd+S |
| Save as | Cmd+Shift+S |
| New tab | Cmd+T |
| Close tab | Cmd+W |
| Force close tab (discard changes) | Cmd+Alt+W |
| Reopen last closed tab | Cmd+Shift+T |
| Next / previous tab | Ctrl+Tab / Ctrl+Shift+Tab |
| Go to tab 1 to 8 | Cmd+1 … Cmd+8 |
| Focus URL bar | Cmd+L |
| Search requests | Cmd+K |
| Find in response | Cmd+F |
| Toggle sidebar | Cmd+\\ |
| Toggle console | Cmd+Alt+C |
| Rename selected item | Cmd+E |
| Duplicate selected item | Cmd+D |
| Import | Cmd+O |
| Settings | Cmd+, |
| Switch environment (Xenon addition) | Cmd+Shift+E |

A Settings page lists all shortcuts; remapping is a nice-to-have after the MVP.

**Environment switcher**

- A dropdown at the top right, always visible, showing the active environment of the current workspace, with "No environment" as an option.
- An eye button next to it opens a quick look at the active environment's variables, editable in place, as in Postman.
- Each workspace remembers its own active environment.
- An environment can be given a colour; the switcher shows it, so prod (for example red) is hard to miss.

## Tech approach

Xenon is a Wails v3 desktop app: a Go backend that does all the real work, and a React UI that only displays and edits. Wails v3 is in beta with a stable API ([release notes](https://github.com/wailsapp/wails/releases)); pin the latest beta and move to 3.0 when it ships.

| Part | Choice |
| --- | --- |
| Desktop shell | Wails v3 (system webview, small binary, generated TypeScript bindings for Go services) |
| UI | React + TypeScript + Vite |
| UI state | Zustand |
| Code editors (body, scripts) | CodeMirror 6, with `pm.*` autocomplete |
| Drag and drop (sidebar, tabs) | dnd-kit |
| HTTP engine | Go `net/http` with `httptrace` for timings, a cookie jar per workspace |
| Scripts | goja (JavaScript engine in pure Go), sandboxed, with a Go-implemented `pm` object and bundled Chai for `pm.expect` |
| Storage | YAML files via `yaml.v3`, one file per request |
| Secrets | Vault file encrypted with age; its key in the OS keychain via `go-keyring` |
| Git sync | go-git (pure Go): clone, pull, push; conflicts resolved per file |
| Postman import/export | Own Go converters for collection v2.1 and environment JSON |

**Rules**

- All logic lives in Go: sending requests, running scripts, reading and writing files, secrets, Git. React never touches the disk or the network directly.
- The Go core is a standalone package (`core/`) with no Wails imports, so a `xenon` CLI (for example `xenon run collection`) can reuse it later.
- Scripts run in goja, never in the webview, so they stay sandboxed and behave the same in the app and in a future CLI.
- One Go service per area (workspaces, requests, environments, vault, sync, import/export), each exposed to React through Wails bindings.

## Build order and open questions

This list is close to full Postman parity, so build it in stages that are each usable on their own, and move real work over as early as stage 3.

1. **Shell and storage:** workspaces, collections, folders, YAML on disk, sidebar, tabs, request builder (params, headers, body, Bearer and Basic auth), response viewer. Simple calls work.
2. **Environments:** variables and scopes, environment switcher, secrets vault.
3. **Postman import:** collections and environments, so real data is in Xenon early and gaps show up fast.
4. **Scripts:** pre-request and post-response, inheritance, the `pm` API, tests, console. This is the largest stage and the one that decides if Postman can be dropped.
5. **Daily comfort:** drag and drop, tab right-click menu, hotkeys, cookies, code snippets panel, remaining auth types (OAuth 2.0, AWS Signature, Digest).
6. **Export:** every level, Xenon YAML and Postman, with encrypted secrets export.
7. **Git sync:** one repo per workspace, sign-in per workspace, commit, push, pull and conflict choices. Until then, the workspace folders can be committed by hand, since they are plain files.

Stages 4 (scripts) and the OAuth 2.0 part of 5 carry most of the risk. Everything else is a lot of UI, but well understood.

**Decisions (Oct 9, 2026)**

- [x] Import from Bruno too? **Yes** — a Bruno (.bru) collection importer is part of stage 3, alongside Postman.
- [x] Platforms? **macOS, Windows and Linux from day one**, built and tested on all three.
- [x] Keep workspace globals as a scope? **Yes** — scopes stay local → environment → collection → globals, and `pm.globals` works.

**Open questions**

- [ ] Open-source Xenon once it replaces Postman, or keep it private? (Not needed for the MVP.)
- [ ] Which scripts in the current Postman collections use parts of `pm` beyond the list in Scripts? A quick scan of the exports would settle the MVP scope.
