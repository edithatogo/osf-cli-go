# GitHub Copilot OSF MCP Package

This package contains repository and workspace MCP configuration for GitHub
Copilot clients that support MCP servers.

The repository-hosted marketplace entry is at `.github/plugin/marketplace.json`.
From a public checkout, install it with:

```text
copilot plugin marketplace add edithatogo/osf-cli-go
copilot plugin install osf-cli-go@osf-cli-go
```

The repository marketplace is available for installation; acceptance into any
GitHub-maintained default marketplace is a separate provider gate.

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

## Install And Validate

Repository configuration:

```powershell
Copy-Item plugins\github-copilot-osf\.github\mcp.json .github\mcp.json
Get-Content .github\mcp.json | ConvertFrom-Json
```

Workspace configuration:

```powershell
Copy-Item plugins\github-copilot-osf\.mcp.json .mcp.json
Get-Content .mcp.json | ConvertFrom-Json
```

For release review or attachment, use the generated
`github-copilot-osf-<version>-<runtime>.zip` archive so the `bin\osf-mcp`
binary is bundled with the MCP JSON files.
