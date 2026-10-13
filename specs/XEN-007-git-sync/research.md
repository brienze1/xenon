# Research: Git Sync (Stage 7)

Verified 2026-10-09 from go-git `releases/v5.x` and `main` sources, COMPATIBILITY.md and issues;
GitHub, GitLab, Atlassian and Microsoft docs; GCM and GitHub Desktop sources.

## R1. Git engine
- **Decision**: `github.com/go-git/go-git/v5@v5.19.3` (latest stable; v6 is beta — migrate after
  GA for autocrlf/filemode/Windows symlink safety). All operations on a repo serialised by a
  per-repo mutex (go-git is not safe for concurrent use).
- **Capabilities used**: clone (`SingleBranch`, `NoCheckout`, full history), checkout with
  `SparseCheckoutDirectories: [subfolder]` (directory sparse), fetch, push with `ForceWithLease`
  pinned to the fetched hash, commit with explicit tree building, `MergeBase`, `IsAncestor`,
  `DiffTree`.
- **Not available → Xenon behaviour**: no merge/rebase (own sync algorithm, R2); no hooks (stated
  in UI); no LFS (link refused if `.gitattributes` has `filter=lfs` for paths under the
  subfolder); no credential helpers (own accounts, R4); no autocrlf in v5 (Xenon writes LF;
  reads CRLF — XEN-001); no `index.lock` (irrelevant: clones are Xenon-only); branch switching in
  sparse mode buggy (#2460) — never switch branches; re-clone on branch change.
- Status scoped to the subfolder; sparse checkout keeps the worktree small.

## R2. Sync algorithm (pull)
1. Fetch the branch. Let L = local HEAD, R = fetched remote, B = merge base.
2. L == R → up to date. L ancestor of R → fast-forward (update ref + checkout, preserving
   uncommitted work by stashing Xenon-side: snapshot uncommitted files, reset, re-apply; a
   re-applied file that the remote also changed → conflict). R ancestor of L → ahead.
3. Diverged → **linear replay**: compute per-path changes B→L (local commits, only under the
   subfolder) and B→R; classify each path:
   - only one side changed → take it; both changed identically → take it;
   - both changed differently → item conflict, except `collection.yaml`/`folder.yaml`/
     `workspace.yaml`/environment rows handled by the semantic merge (R3);
   - delete vs modify → conflict (Keep / Delete).
4. Build the result tree = R's tree with the merged subfolder subtree; create **one commit** on
   top of R ("Sync: <local commit subjects>", author = account) — local commits are squashed into
   one replayed commit (linear history, no merge commits), then reset local branch to it.
5. With conflicts: write nothing to the branch; open the Conflict view on in-memory versions; on
   Finish build the tree with the resolutions and commit as in 4; on Cancel nothing changed.
- **Rationale**: Xenon owns every commit in its clone, so replaying is safe; one file per request
  keeps conflicts per item (PRD).

## R3. Semantic merge for structural files
- `collection.yaml` / `folder.yaml` / `workspace.yaml`: parse base/mine/theirs; merge field by
  field (XEN-001 node patcher); `order`: theirs order, then my additions not in theirs, minus
  entries deleted on either side; scalar fields changed differently on both sides → field-level
  conflict (Keep mine / Keep theirs).
- Environment files: rows merged by key (3-way per row); a row changed differently on both sides
  → the environment is an item conflict (Keep mine / theirs / both) as the spec requires.
- Unknown fields: 3-way at node level; differing → take theirs and note in the summary.

## R4. Accounts and transport auth
- **GitHub device flow** (OAuth App, client id only; `repo` + `read:user` + `user:email` +
  `write:public_key` scopes); poll per `interval`/`slow_down`, 15-minute code expiry. Token sent
  over HTTPS as BasicAuth (`x-access-token`/any username + token). Client id provided by the
  owner (spec Assumptions) via build-time variable `XENON_GITHUB_CLIENT_ID`.
- **Token accounts**: BasicAuth with host conventions — GitHub any username; GitLab `oauth2`;
  Bitbucket `x-bitbucket-api-token-auth` (API tokens; app passwords removed 2026-07-28); Azure
  DevOps empty username (org-scoped PATs; global PATs end 2026-12-01). Verified on save with a
  `ls-remote`-style advertisement request (or the host's user API when known).
- **SSH**: `NewSSHAgentAuth` (Unix socket; Windows Pageant / OpenSSH named pipe),
  `NewPublicKeysFromFile` with passphrase prompt, generated ed25519 (`ssh.MarshalPrivateKey…`,
  0600, in app data `ssh/`); `~/.ssh/config` not used.
- **Host keys**: Xenon's own `known_hosts` in app data, pre-seeded with GitHub, GitLab.com and
  Bitbucket published keys; TOFU prompt for others; mismatch blocks.
- **Secrets storage**: tokens and key passphrases in the OS keychain (`go-keyring`, service
  `dev.brienze.xenon.git`), passphrase fallback via XEN-002's vault mechanism (app-level vault).
- **Org restrictions**: GitHub OAuth App restrictions / SAML SSO detected from 403 messages →
  guidance (spec FR-019).

## R5. Testing
- Local bare repos (`git init --bare` via go-git) served over `file://`, an in-process smart-HTTP
  server (go-git `plumbing/transport/server` or `git http-backend` when available) and an in-process
  SSH server (`gliderlabs/ssh` + go-git upload/receive-pack) for transport/auth tests; a mock
  GitHub device-flow endpoint (httptest) behind the GitHub API base URL property; two Xenon clones
  in BDD for conflict scenarios; property `XENON_GITHUB_API_URL` / `XENON_GITHUB_LOGIN_URL` for
  tests.
