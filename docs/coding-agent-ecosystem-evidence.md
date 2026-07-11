# Coding Agent Ecosystem Evidence

Last reviewed: 2026-07-11

## Authoritative references

- [Cursor MCP](https://docs.cursor.com/context/model-context-protocol)
- [Cline MCP configuration](https://www.mintlify.com/cline/cline/mcp/adding-and-configuring-servers)
- [Roo Code MCP](https://roocodeinc.github.io/Roo-Code/features/mcp/using-mcp-in-roo/)
- [VS Code MCP servers](https://code.visualstudio.com/docs/agent-customization/mcp-servers)
- [VS Code MCP configuration reference](https://code.visualstudio.com/docs/agents/reference/mcp-configuration)
- [Zed MCP](https://zed.dev/docs/ai/mcp)
- [Zed MCP server extensions](https://zed.dev/docs/extensions/mcp-extensions)

## Repository artifacts

The repository now provides standard MCP configuration templates for:

- Cursor: `.cursor/mcp.json`
- Cline: `integrations/cline/cline_mcp_settings.json`
- Roo Code: `.roo/mcp.json`
- Windsurf: `integrations/windsurf/mcp_config.json`
- VS Code: `.vscode/mcp.json`
- Zed: `integrations/zed/settings.json`

Each template launches `go run ./cmd/osf-mcp` and references `OSF_TOKEN`,
`OSF_USERNAME`, and `OSF_PASSWORD` through environment substitution. The
templates contain no credentials.

## Installation surfaces

Copy the client-specific template to the path described in
`integrations/README.md`, or use the client’s MCP add-server UI. VS Code also
supports workspace `.vscode/mcp.json`; Cursor supports project `.cursor/mcp.json`;
Roo Code supports project `.roo/mcp.json`.

Zed’s native MCP extension route requires a compiled Zed extension and is being
deprecated in favor of the official MCP registry. This repository supplies the
documented direct `context_servers` configuration instead of claiming a Zed
extension listing.

## Validation and external status

```text
go run ./tools/checkreleasecontract
```

The release contract parses every template and checks its command, arguments,
and credential environment references. Provider galleries, one-click
collections, and marketplace listings are not called submitted or approved
without dated provider evidence.
