# OSF CLI Go for Codex

This plugin exposes the read-only OSF MCP server to Codex-compatible hosts.

The release archive bundles the server at `bin/osf-mcp` and loads the MCP
configuration from `.mcp.json`. For source development from the repository
root, run:

```text
go run ./cmd/osf-mcp
```

Set `OSF_TOKEN` before using authenticated tools. `OSF_USERNAME` and
`OSF_PASSWORD` are supported as a fallback. The plugin manifest can be checked
with:

```text
claude plugin validate plugins/codex-osf
```
