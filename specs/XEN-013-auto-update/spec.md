# Feature Specification: Auto-Update Xenon From GitHub Releases

**Feature Branch**: `xen-013`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "XEN-013 — Xenon updates itself from the builds published as GitHub Releases on brienze1/xenon. Every merge to main publishes a release with an automatically bumped version, portable packages for macOS (ad-hoc signed), Windows and Linux, and checksums. Settings → About/Updates shows the running and latest version with Check for updates; when behind, Update downloads the matching package, verifies it, replaces the installed app and restarts — on all three OSes. An 'Update automatically' setting (per machine) checks at every startup without blocking it, downloads and verifies in the background, then asks before restarting through a bottom-left toast (Restart now / Later) and a persistent footer notice. Offline, unreachable or rate-limited checks fail quietly with 'Couldn't check for updates'. Go owns all the logic; React only displays. Source: GitHub issue brienze1/xenon#14."

## Context

Today a new Xenon build reaches a machine by hand: `make package`, quit Xenon, replace
`~/Applications/Xenon.app`. Verified against the code at `25acd4d`:

- The repository has **no GitHub Releases** and `.github/workflows/ci.yml` has **no release
  job** — nothing publishes a build.
- The version is hard-coded in three places that nothing keeps in step: `build/config.yml`
  (`info.version: "0.1.0"`), the macOS `Info.plist` (`CFBundleShortVersionString` 0.1.0) and the
  Go default (`defaultVersion = "0.1.0"` in `internal/infra/properties`, overridable with
  `XENON_VERSION`). The running version is already logged at startup and used in export headers
  and the MCP server name.
- macOS builds are **ad-hoc code-signed** (`codesign --sign -`), not notarized.
- Machine-wide app settings (script timeout, History caps, AI access) already live in the
  per-machine app registry in Xenon's app-data folder and are edited in the Settings tab.
- Open tabs, including their unsaved drafts, are already persisted in the UI state and restored
  after a restart.
- The app footer has a status area on its left; the app has no toast component yet.

## Clarifications

### Session 2026-10-15 (before specification)

- Q: What publishes a release and where does the version come from? → A: **Every merge to
  `main`** publishes a release; the patch number is bumped automatically per merge.
- Q: Which OSes install updates by themselves in this first version? → A: **All three** —
  macOS replaces the `.app` bundle, Windows the `.exe`, Linux the executable.
- Q: Is ad-hoc signing enough on macOS? → A: **Yes.** No Developer ID, no notarization. The
  release notes explain that the first manual download from a browser needs right-click → Open
  (or clearing the quarantine attribute); updates Xenon downloads itself are not quarantined.
- Q: With "Update automatically" on and a newer version found at startup, what happens? → A:
  Download and verify in the background, then **ask before restarting**: a toast in the
  bottom-left with **Restart now** and **Later**, and a persistent notice in the footer showing
  an update is ready, with a restart action, until Xenon restarts.
- Q: Patches are bumped per merge; how is a minor or major bump requested? → A: A `VERSION`
  file at the repository root holds `major.minor` (starting at `0.1`). A release takes the next
  patch within that line (`.0` when the line has no release yet); a pull request that changes
  `VERSION` to `0.2` makes the next release `v0.2.0`.
- Q: What if the installed app's location is not writable? → A: **Refuse with a message** — no
  admin prompt. The section names the folder, suggests moving Xenon to a folder the user owns
  (e.g. `~/Applications`) and links the release page for a manual install.
- Q: After Later, when is the verified update installed? → A: **On the next normal quit.**
  Quitting swaps in the verified update, so the next launch runs the new version.

### Session 2026-10-15 (speckit-clarify)

- Q: What does Update automatically default to? → A: **On**, for fresh and existing installs
  (a machine with no stored value reads it as on). The restart is still always asked.
- Q: With Update automatically off, does Xenon still check at startup? → A: **Check and
  notify** — it checks once at startup and, when an update is offered, shows an "Update
  available" footer notice that opens Settings; nothing is downloaded until the user clicks
  Update.
- Settled from the code, not asked: update activity (check result, download, verification,
  install, failures) is logged through the existing app log (`<app-data>/logs/xenon.log`) like
  every other operation; nothing new is shown beyond the section, toast and footer.
- Settled, not asked: the checksums file comes from the same release over HTTPS, so
  verification guards against corrupt or truncated downloads, not against a compromised GitHub
  account — the same trust the user already places in the repository; packages are downloaded
  only from the URLs the release itself lists.
- Settled, not asked: a downloaded update that was not installed (Xenon crashed, or was killed
  rather than quit) is discarded at the next start; the startup check simply finds it again.
- Settled, not asked: the section links the latest release's GitHub page ("What's new") instead
  of rendering release notes inside Xenon.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See whether Xenon is up to date (Priority: P1)

Luis opens Settings and finds an **About & Updates** section. It shows the version he is running
and, after he clicks **Check for updates**, the latest published version and whether he is up to
date.

**Why this priority**: Knowing the running version and whether it is behind is the base every
other story builds on, and on its own it already replaces guessing.

**Independent Test**: With a faked releases source publishing a higher version, open Settings,
click Check for updates and read both versions and the "update available" status.

**Acceptance Scenarios**:

1. **Given** Xenon 0.1.4 is running, **When** Luis opens Settings, **Then** the section shows
   "Version 0.1.4".
2. **Given** the latest release is 0.1.7, **When** he clicks Check for updates, **Then** the
   section shows "Latest version 0.1.7", says an update is available and shows an **Update**
   button.
3. **Given** the latest release is 0.1.4 (or older), **When** he checks, **Then** the section
   says "Xenon is up to date" and shows no Update button.
4. **Given** there is no network, GitHub cannot be reached, answers with an error or is
   rate-limiting, **When** he checks, **Then** the section says "Couldn't check for updates",
   keeps the running version visible, offers Check for updates again, and nothing else in the app
   is affected.
5. **Given** a check is running, **When** he looks at the section, **Then** it shows
   "Checking…" and the Check for updates button is disabled until the check ends.

---

### User Story 2 - Update with one click (Priority: P1)

Luis is behind. He clicks **Update**; Xenon downloads the package for his OS and processor,
verifies it, replaces the installed app and restarts on the new version, with his tabs and
unsaved drafts back where they were.

**Why this priority**: This is the feature's purpose — replacing the manual build-and-copy.

**Independent Test**: With a faked release carrying a package and its checksum, click Update
and confirm the installed app files are replaced by the package's content and a restart is
requested; with a wrong checksum, confirm nothing is replaced.

**Acceptance Scenarios**:

1. **Given** an update is available, **When** Luis clicks Update, **Then** the section shows
   download progress, then "Installing…", the installed app is replaced by the new version and
   Xenon restarts on it.
2. **Given** the downloaded package does not match its published checksum, **When** the update
   runs, **Then** the installed app is left untouched, the downloaded file is discarded and the
   section says "The update could not be verified — nothing was changed".
3. **Given** the release has no package for this OS and processor, **When** Luis checks,
   **Then** the section shows the latest version, says "No update package for this system" and
   shows no Update button.
4. **Given** the download fails part-way (network drop), **When** it fails, **Then** the
   installed app is untouched and the section says "The update could not be downloaded" with
   Update offered again.
5. **Given** the installed app sits where Xenon cannot write, **When** the update would
   replace it, **Then** nothing is changed, no admin prompt appears, and the section says "Xenon can't
   replace itself at <folder>" with a suggestion to move it to a folder Luis owns (such as
   `~/Applications`) and a link to the release page for a manual install.
6. **Given** Xenon restarted after an update, **When** it opens, **Then** the tabs, drafts and
   layout are the same as before the restart and Settings shows the new version.

---

### User Story 3 - Updates arrive by themselves (Priority: P2)

Luis ticks **Update automatically**. From then on, every time Xenon starts it checks for a newer
version in the background; when one exists it downloads and verifies it without blocking his
work, then a toast in the bottom-left says the update is ready with **Restart now** and
**Later**, and the footer keeps an "Update ready — Restart" notice until he restarts.

**Why this priority**: Saves the trip to Settings, but depends on Stories 1 and 2.

**Independent Test**: With the setting on and a faked newer release, start the app and confirm
the startup is not delayed, the package is downloaded and verified, and an "update ready" event
reaches the UI; with the setting off, confirm nothing is downloaded.

**Acceptance Scenarios**:

1. **Given** Update automatically is on and a newer release exists, **When** Xenon starts,
   **Then** the window opens and works immediately, the update downloads and is verified in the
   background, and a bottom-left toast offers Restart now and Later.
2. **Given** the toast is shown, **When** Luis clicks Restart now, **Then** Xenon installs the
   update and restarts on it.
3. **Given** the toast is shown, **When** Luis clicks Later (or dismisses the toast), **Then**
   the toast closes and the footer shows "Update ready — Restart"; when Luis later quits Xenon
   normally, the verified update is installed during the quit and the next launch runs the new
   version.
4. **Given** the footer notice is shown, **When** Luis clicks its Restart action, **Then**
   Xenon installs the update and restarts on it.
5. **Given** Update automatically is off and a newer release exists, **When** Xenon starts,
   **Then** it checks once in the background, downloads nothing, and the footer shows "Update
   available" which opens the Settings section when clicked.
6. **Given** a machine that never stored the setting (fresh install or an existing install
   upgrading into this feature), **When** Settings opens, **Then** Update automatically is
   ticked.
7. **Given** Update automatically is on and the startup check fails (offline, rate-limited),
   **When** Xenon starts, **Then** it opens normally, shows no toast and no error dialog, and the
   Settings section reads "Couldn't check for updates".
8. **Given** Luis ticks or unticks Update automatically, **When** he restarts Xenon, **Then**
   the setting is as he left it, on this machine only.

---

### User Story 4 - Every merge publishes a release (Priority: P1)

A pull request is merged into `main`. CI builds Xenon for macOS, Windows and Linux with the next
version embedded, and publishes a GitHub Release with the packages and a checksums file.

**Why this priority**: Without published releases there is nothing to update to; Stories 1–3
cannot work in the real world without it.

**Independent Test**: Merge a change to `main` and see a new release on GitHub whose version is
one patch above the previous release, with one package per supported OS and a checksums file,
and whose packaged app reports that version in Settings.

**Acceptance Scenarios**:

1. **Given** the latest release is 0.1.6, **When** a pull request merges into `main`, **Then**
   a release `v0.1.7` is published with the macOS, Windows and Linux packages and a checksums
   file listing every package.
2. **Given** a release is being built, **When** any OS build fails, **Then** no release is
   published for that merge (no partial release).
3. **Given** a published release, **When** its macOS package is installed and opened, **Then**
   Settings shows the release's version.
4. **Given** the maintainer wants a new minor or major version, **When** they merge a pull
   request changing the `VERSION` file from `0.1` to `0.2`, **Then** that merge publishes
   `v0.2.0` and the following merges publish `v0.2.1`, `v0.2.2`, …

### Edge Cases

- The latest release is a draft or prerelease → ignored; only published, non-prerelease
  releases count.
- The running build is newer than the latest release (a local dev build) → "Xenon is up to
  date"; never offer a downgrade.
- A development run (not a packaged app, e.g. `wails3 dev` or the BDD harness) → checking works,
  but installing is refused with "Updates install only in a packaged Xenon".
- A second Check for updates or Update click while one is running → ignored; one operation at a
  time.
- The release lists a package but the checksums file is missing or has no line for it →
  treated as unverifiable; nothing is installed.
- Xenon quits or crashes mid-download → the partial file is discarded; the installed app is
  untouched.
- Replacing the app fails half-way (disk full, file locked) → the previous app is restored and
  the section reports the failure.
- On Windows the running `.exe` cannot be overwritten in place → the update must still succeed
  (the running file is moved aside and removed on the next start).
- The release's version string is not a valid version → the check reports "Couldn't check for
  updates".
- Several Xenon windows/processes → only one app instance runs; this feature does not change
  that. An AI client's `xenon mcp` bridge process started from the old executable keeps running
  the old version until the AI client restarts it; on Windows it may keep the old `.exe` locked,
  so removing it is retried at each start until it succeeds.
- macOS runs a quarantined app straight from a download location in a read-only translocated
  folder → reported as the "can't replace itself" case, with the move-to-`~/Applications`
  suggestion.
- The disk fills up during the download → "The update could not be downloaded"; the partial file
  is removed.
- The toast stays until the user chooses Restart now, Later or closes it; it never disappears by
  itself.

## Requirements *(mandatory)*

### Functional Requirements

**Versioning and publishing**

- **FR-001**: Every merge to `main` MUST publish a GitHub Release. Its version is
  `<major.minor>.<patch>`, where `major.minor` comes from the repository's `VERSION` file and
  `patch` is one more than the highest released patch in that line, or `0` when the line has no
  release yet (so the first release is `v0.1.0`).
- **FR-002**: The version MUST have a single source of truth from which the packaged app, the
  macOS bundle metadata and the release tag all take it; a running Xenon MUST be able to report
  its own version.
- **FR-003**: Each release MUST carry one package per supported system — macOS (one package
  running on Apple silicon and Intel), Windows x64 and Linux x64 — and a checksums file with a
  SHA-256 digest for every package.
- **FR-004**: A release MUST be published only when every package built; a failed build
  publishes nothing.
- **FR-005**: The macOS package MUST be ad-hoc signed, and the release notes MUST state how to
  open a manually downloaded first copy (right-click → Open, or clearing the quarantine
  attribute).

**Checking**

- **FR-006**: Settings MUST show an About & Updates section with the running version, the latest
  published version (once known), the update status and a Check for updates action.
- **FR-007**: Checking MUST ask GitHub for the latest published, non-draft, non-prerelease
  release of `brienze1/xenon` and compare its version with the running version by semantic
  version order.
- **FR-008**: An update MUST be offered only when the latest release is strictly newer than the
  running version and has a package for the running OS and processor.
- **FR-009**: A check that cannot complete (no network, unreachable, error answer, rate limit,
  unreadable answer) MUST end in the status "Couldn't check for updates", MUST NOT show an error
  dialog and MUST NOT block startup or any other work.
- **FR-010**: Only one check or update MUST run at a time; a request while one runs MUST be
  ignored.

**Updating**

- **FR-011**: Update MUST download the package for the running OS and processor, report download
  progress, and verify it against the release's checksums file before anything is installed.
- **FR-012**: A package that fails verification, or has no checksum, MUST be discarded and MUST
  NOT change the installed app.
- **FR-013**: Installing MUST replace the installed app as a whole — the `.app` bundle on macOS,
  the `.exe` on Windows, the executable on Linux — so that a failure at any point leaves the
  previous app working.
- **FR-014**: After installing, Xenon MUST restart on the new version; the open tabs, drafts and
  layout MUST be restored as on any restart.
- **FR-015**: When the installed app's location is not writable, Xenon MUST NOT install or ask
  for admin rights; the status MUST read "Xenon can't replace itself at <folder>", suggest moving
  Xenon to a folder the user owns and link the release page.
- **FR-016**: A development run that is not a packaged app MUST refuse to install, with the
  status "Updates install only in a packaged Xenon".

**Automatic updates**

- **FR-017**: Settings MUST offer an **Update automatically** option, stored per machine with the
  other app settings, **on by default** (a machine with no stored value reads it as on).
- **FR-018**: Xenon MUST check once at every startup, after the window is usable, whatever the
  setting. With Update automatically on, when an update is offered it MUST download and verify
  it in the background.
- **FR-018a**: With Update automatically off, when the startup check offers an update Xenon MUST
  download nothing. Whenever the state is "update available" (after the startup check or a
  manual check, whatever the setting), the footer MUST show an "Update available" notice that
  opens the Settings tab, whose first section is About & Updates.
- **FR-018b**: When a background (automatic) download or verification fails, Xenon MUST show no
  toast and no dialog; the About & Updates section shows the failure, and the next startup tries
  again.
- **FR-019**: When an automatically downloaded update is verified, Xenon MUST show a toast in the
  bottom-left of the window with **Restart now** and **Later**, and a footer notice "Update
  ready — Restart" that stays until Xenon restarts.
- **FR-020**: Restart now (toast) and Restart (footer) MUST install the verified update and
  restart; Later MUST close the toast and leave the footer notice, and the verified update MUST then be
  installed when Xenon is next quit normally.
- **FR-021**: A downloaded update that was not installed before Xenon stopped MUST be discarded
  at the next start.
- **FR-022**: Packages MUST be downloaded only from the locations listed by the release itself.
- **FR-023**: Update activity — check results, downloads, verification, installation and failures
  — MUST be written to the app log, never including anything but versions, file names, sizes and
  statuses.
- **FR-024**: When the latest version is known, the section MUST link its release page
  ("What's new").

### Key Entities *(include if feature involves data)*

- **Release**: a published version of Xenon — version, release page, the packages it carries
  (one per OS and processor, each with a download location and size) and its checksums file.
- **Update status**: what the section and footer show — idle, checking, up to date, update
  available, no package for this system, couldn't check, downloading (with progress), ready to
  install, installing, failed (with a reason) — plus the running and latest versions.
- **App settings**: the existing per-machine settings, gaining "update automatically".

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every merge to `main` produces a published release with packages for all three
  systems and a checksums file, with no manual step.
- **SC-002**: Going from an older installed version to the latest takes at most two clicks
  (Check for updates, Update) and no file handling by the user.
- **SC-003**: Whatever the Update automatically setting, the window becomes usable as fast with an update
  pending as without one (the check adds no waiting at startup).
- **SC-004**: In 100% of the failure cases listed (offline, rate limit, bad checksum, failed
  download, failed replacement), the previously installed app still opens afterwards.
- **SC-005**: After an update restart, 100% of open tabs and unsaved drafts are restored.

## Assumptions

- GitHub's public, unauthenticated releases API is used; its per-address limit (60 requests an
  hour) is far above one check per startup plus manual checks, and hitting it is just a
  "Couldn't check" case.
- Supported update targets are macOS (Apple silicon and Intel in one package), Windows x64 and
  Linux x64 — the systems the existing CI runners build. Other processors get "No update package
  for this system".
- The Windows and Linux packages are the portable executable, not an installer; a copy installed
  through a system package manager is outside this feature. The Linux executable needs the same
  system libraries the existing Linux build needs (GTK 4 and WebKitGTK 6); Windows needs the
  WebView2 runtime that ships with Windows 10/11.
- Automatic checks happen only at startup; there is no periodic re-check while Xenon runs.
- Release notes are generated from the merged changes; their wording is not specified here.
