# Coding Agent MCP Integrations

These templates expose the OSF MCP server through the standard configuration
surface documented by each client. They use `go run ./cmd/osf-mcp` for local
development and read credentials from environment variables; no secrets are
stored in the repository.

| Client | Template | Installation or copy target | Provider status |
| --- | --- | --- | --- |
| Cursor | `.cursor/mcp.json` | Project `.cursor/mcp.json` or `~/.cursor/mcp.json` | MCP configuration available; Cursor collection listing not claimed |
| Cline | `integrations/cline/cline_mcp_settings.json` | `~/.cline/data/settings/cline_mcp_settings.json` | MCP configuration available; Cline Marketplace listing not claimed |
| Roo Code | `.roo/mcp.json` | Project `.roo/mcp.json` | MCP configuration available; Roo marketplace listing not claimed |
| Windsurf | `integrations/windsurf/mcp_config.json` | Windsurf MCP configuration file | MCP configuration available; provider listing not claimed |
| VS Code | `.vscode/mcp.json` | Workspace `.vscode/mcp.json` | MCP configuration available; VS Code gallery submission not claimed |
| Zed | `integrations/zed/settings.json` | Merge `context_servers` into Zed settings | Direct MCP configuration available; native MCP extension route is deprecated in favor of the official registry |

## Validation

The repository release contract parses every template and verifies the `osf`
server uses the expected stdio command without embedded credentials:

```text
go run ./tools/checkreleasecontract
```

Provider galleries and one-click collections require separate provider-side
review or indexing. This repository does not describe a direct integration as
submitted or approved without dated provider evidence.
