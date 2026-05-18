# OSF CLI Go for Claude

This plugin exposes the OSF MCP server to Claude Code and Claude/Cowork.

Expected bundled binary path:

```text
bin/osf-mcp
```

For local development from the repository root, use:

```powershell
go run ./cmd/osf-mcp
```

Set `OSF_TOKEN` before using authenticated tools. `OSF_USERNAME` and
`OSF_PASSWORD` are accepted as a fallback.
