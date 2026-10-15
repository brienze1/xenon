# Quickstart: validating request aliases

## Automated

```bash
make bdd TAG=@aliases         # every scenario in test/integration/features/aliases/ and the alias scenarios elsewhere
make bdd                      # the whole Godog suite stays green
cd frontend && npm test -- AliasMenu use-tab-title ItemBreadcrumb HistoryView request-builder/store tabs/store
make frontend-check           # lint + typecheck + every Vitest spec
make validate                 # final gate
```

## Manual (in `wails3 dev`)

1. Open a saved request **Login**. Top-right of the tab: `Default ▾`. Open it: `Default`,
   `Create new alias…`.
2. Create `mx-test-user`. The tab title becomes `Login · mx-test-user`, the breadcrumb ends with
   `mx-test-user`, the content equals the saved Default. The sidebar still has one **Login** row.
3. Change the body, Cmd+S, Send: the request sent carries the new body. Open the file: an
   `aliases:` list with `mx-test-user`; the Default's body is unchanged.
4. Switch to Default, edit the URL (do not save), switch back: no prompt; each version shows its own
   edits; the tab has the dirty dot. Close the tab: Save / Don't Save / Cancel; Save writes both.
5. Try to create `default`, `MX-TEST-USER ` and an empty name: each is refused inline.
6. Create `test-user`, click its crumb, rename to `qa-user`, Cmd+S: dropdown, crumb and title update.
7. `⋯` → Delete… on `qa-user`, confirm: the tab returns to Default and the file keeps only
   `mx-test-user`. On another request, create one alias and delete it: that file has no `aliases:` key
   and is byte-identical to before.
8. Create `canary` on **Orders**, quit, restart: the tab is restored on `canary`. Click **Orders** in
   the sidebar: same tab, still `canary`. Close it, open again: `Default`.
9. History shows `Login · mx-test-user` for the alias send.
10. Export the collection to Postman: `Login` and `Login (mx-test-user)` are both there.
11. Give `mx-test-user` a post-response script `pm.environment.set('mxToken', …)` and its own OAuth 2.0
    auth: sending it sets only `mxToken`, and its token is separate from the Default's.
12. With an AI client: `tree` shows `aliases`; `tree_edit` `kind: alias` `create`; `send` with
    `alias_id`.
