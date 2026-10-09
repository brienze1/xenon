# Quickstart: validating stage 7

## Automated
```bash
make bdd TAG=@xen_007
go test ./test/integration/property/...   # merge engine property test
make validate
```

## Manual (≈20 min, `make dev`) — needs a test GitHub repo and the OAuth App client id
1. Settings → Git accounts → Add GitHub account → device code → approve in the browser → account
   listed. Add a GitLab token account and an SSH key (generate; upload to GitHub).
2. Link "Acme" to `github.com/<you>/xenon-sync-test`, branch `main`, subfolder `api/xenon` → clone
   created under app data; footer "Up to date" after first commit+push.
3. Edit two requests → Sync panel lists them → uncheck one → Commit and Push → GitHub shows only
   the checked file under `api/xenon/`; no secret values in the commit.
4. On GitHub (web editor) change a different request → footer "Behind 1" → Pull → change appears;
   open clean tab reloads.
5. Change the same request locally and on GitHub → Pull → Conflict view → Keep both → two
   requests; Finish → Push.
6. Turn on auto-push → save → pushed within ~3 s.
7. Kill the network → Push → "Offline — will retry"; restore → Push works.
8. Clone the repo as a new workspace on another machine → secrets show missing.
