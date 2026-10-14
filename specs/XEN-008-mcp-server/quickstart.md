# Quickstart: validating XEN-008

## Automated

```bash
make bdd TAG=@xen_008      # every XEN-008 feature (tags: @all @xen_008 @<area>)
make bdd TAG=@mcp          # test/integration/features/mcp/*.feature
make bdd TAG=@ai_activity  # test/integration/features/history/ai_activity.feature
make bdd TAG=@app_settings # test/integration/features/settings/app_settings.feature
make bdd                   # full suite stays green
cd frontend && npm run lint && npm run typecheck && npm test
make validate
```

Expected: all green; `leak_guard.feature` reports 0 plaintext hits (SC-004); `connect.feature` asserts 24 tools,
hints and the tools/list size budget (SC-005, SC-007).

## Manual (real client)

1. `make dev` (or the built app). Open a workspace with a collection.
2. Click the gear (or Cmd/Ctrl+,): a **Settings tab** opens (not a dialog). AI access (MCP) shows
   `Running at http://127.0.0.1:47821/mcp`, the stdio command and both configs.
3. Claude Code: `claude mcp add --transport http xenon http://127.0.0.1:47821/mcp` → `/mcp` lists 24 xenon tools.
4. Ask: "list my collections, then send the Get user request" → status/body come back; History (filter **AI**) shows
   the `tree.list`, `tree.get` action rows and the AI-tagged send.
5. Ask: "create a Payments collection with an Orders folder and a POST /orders request" → items appear in the sidebar
   without reload.
6. Ask: "add a secret apiToken = abc123xyz to Staging and send GET https://postman-echo.com/headers with
   Authorization: Bearer {{apiToken}}" → the send succeeds; the AI's transcript shows `•••••` where the token would be
   echoed.
7. Mark Staging "Confirm before send", ask for a send → Xenon comes to the front with Approve/Deny; Deny → the AI
   reports it was denied. Tick **Dangerously skip approvals**, Save, ask again → sent with no prompt.
8. Turn AI access off, Save → `/mcp` shows xenon disconnected; on again → reconnects.
9. stdio: `claude mcp add xenon -- "<path shown in Settings>" mcp` → same tools. Quit Xenon → the bridge reports
   "Xenon is not running".
10. Settings → History: set User 10,000 and AI 100, Save; make >100 AI calls → only the newest 100 AI rows remain,
    no user send is dropped.
