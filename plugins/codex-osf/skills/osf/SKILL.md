# OSF MCP Tools

Use the `osf` MCP server for read-only Open Science Framework tasks:

- identify the authenticated OSF account;
- list and inspect projects and components;
- list OSF Storage files and folders;
- list project contributors.

Prefer `OSF_TOKEN` for authentication. Use `OSF_USERNAME` and `OSF_PASSWORD`
only when a token is unavailable.

## Install And Validate

From the repository root:

```powershell
    codex plugin marketplace add .
Get-Content plugins\codex-osf\.codex-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\codex-osf\.mcp.json | ConvertFrom-Json
```

For packaged installs, use a generated `codex-osf-<version>-<runtime>.zip`
archive that includes the `bin\osf-mcp` or `bin\osf-mcp.exe` binary.
