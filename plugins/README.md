# Client Plugin Packages

This directory contains client-specific packaging for the OSF MCP server.

## Binary Contract

Packaged plugins expect a built MCP server binary at:

```text
bin/osf-mcp
bin/osf-mcp.exe
```

Local development can use `go run ./cmd/osf-mcp`, but release packages should
bundle the binary or use the MCPB artifacts from `dist/mcpb`.

## Packages

- `claude-osf`: Claude Code/Cowork plugin metadata.
- `github-copilot-osf`: GitHub Copilot MCP configuration package.
- `codex-osf`: Codex plugin metadata and skill.
- `gemini-osf`: Gemini CLI extension metadata.
- `qwen-osf`: Qwen Code extension metadata.

## Validation

```powershell
Get-Content plugins\claude-osf\.claude-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\github-copilot-osf\.github\mcp.json | ConvertFrom-Json
Get-Content plugins\github-copilot-osf\.mcp.json | ConvertFrom-Json
Get-Content plugins\codex-osf\.codex-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\gemini-osf\gemini-extension.json | ConvertFrom-Json
Get-Content plugins\qwen-osf\qwen-extension.json | ConvertFrom-Json
```

Register the local Codex plugin marketplace:

```powershell
codex plugin marketplace add C:\Users\60217257\repos\osf-cli-go\.agents\plugins
```

Build self-contained plugin archives:

```powershell
.\scripts\build-plugin-archives.ps1
```

The `Plugin Archives` GitHub Actions workflow builds platform-specific ZIPs for
manual marketplace review or release attachment.
