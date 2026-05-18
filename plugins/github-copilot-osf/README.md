# GitHub Copilot OSF MCP Package

This package contains repository and workspace MCP configuration for GitHub
Copilot clients that support MCP servers.

## Repository Configuration

Use `.github/mcp.json` for Copilot coding agent repository configuration.

## Workspace Configuration

Use `.mcp.json` with GitHub Copilot CLI or MCP-compatible local clients.

Both configurations start the OSF MCP server with:

```powershell
go run ./cmd/osf-mcp
```

Set `OSF_TOKEN` before starting the client. `OSF_USERNAME` and `OSF_PASSWORD`
are supported as a fallback, but token authentication is preferred.
