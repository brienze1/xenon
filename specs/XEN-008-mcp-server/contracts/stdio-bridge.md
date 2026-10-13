# Contract: `xenon mcp` (stdio bridge)

**Invocation**: `<path-to-xenon-executable> mcp` (no other arguments). On macOS the executable is
`Xenon.app/Contents/MacOS/xenon`; Settings shows the exact absolute path (`os.Executable()`).

**Behaviour**

1. Reads `<DataDir>/app.json` → `settings.aiAccess` (`XENON_DATA_DIR` honoured, same default as the app). Never writes.
2. `enabled == false` → stderr `Xenon AI access is off — turn it on in Xenon → Settings → AI access (MCP).`, exit 1.
3. Connects to `http://127.0.0.1:<port>/mcp` (`XENON_MCP_PORT` honoured). Connection refused / timeout (3 s) → stderr
   `Xenon is not running — start Xenon with AI access on, then retry.`, exit 1.
4. After the client's `initialize`, checks that the server answered with `serverInfo.name == "xenon"`. If it did
   not, prints `Port <n> is in use by another program — change the port in Xenon → Settings → AI access (MCP).` to
   stderr and exits 1, so AI traffic is never forwarded to a foreign server.
5. Pumps newline-delimited JSON-RPC between stdin/stdout and the server session until stdin closes (exit 0) or the
   server goes away (stderr `Xenon closed the connection.`, exit 1). stdout carries only protocol messages.
6. Never starts the Xenon GUI and never triggers single-instance activation.

**Client configs shown in Settings**

```json
{ "mcpServers": { "xenon": { "type": "http", "url": "http://127.0.0.1:47821/mcp" } } }
```

```json
{ "mcpServers": { "xenon": { "command": "/Applications/Xenon.app/Contents/MacOS/xenon", "args": ["mcp"] } } }
```

(port and path substituted with the live values).
