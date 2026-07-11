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

Validate package metadata from the repository root:

```powershell
Get-Content plugins\claude-osf\.claude-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\github-copilot-osf\.github\mcp.json | ConvertFrom-Json
Get-Content plugins\github-copilot-osf\.mcp.json | ConvertFrom-Json
Get-Content plugins\github-copilot-osf\plugin.json | ConvertFrom-Json
Get-Content plugins\codex-osf\.codex-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\gemini-osf\gemini-extension.json | ConvertFrom-Json
Get-Content plugins\qwen-osf\qwen-extension.json | ConvertFrom-Json
Get-Content .github\plugin\marketplace.json | ConvertFrom-Json
Get-Content qwen-extension.json | ConvertFrom-Json
```

Register the local Codex plugin marketplace:

```powershell
codex plugin marketplace add C:\Users\60217257\repos\osf-cli-go\.agents\plugins
```

## Install Paths

Claude Code/Cowork:

```powershell
Copy-Item -Recurse plugins\claude-osf $env:USERPROFILE\.claude\plugins\osf-cli-go
claude plugin validate $env:USERPROFILE\.claude\plugins\osf-cli-go
```

GitHub Copilot:

```powershell
copilot plugin marketplace add edithatogo/osf-cli-go
copilot plugin marketplace browse osf-cli-go
copilot plugin install osf-cli-go@osf-cli-go
copilot plugin install edithatogo/osf-cli-go:plugins/github-copilot-osf
Copy-Item plugins\github-copilot-osf\.github\mcp.json .github\mcp.json
Copy-Item plugins\github-copilot-osf\.mcp.json .mcp.json
Get-Content .github\mcp.json | ConvertFrom-Json
Get-Content .mcp.json | ConvertFrom-Json
```

Codex:

```powershell
codex plugin marketplace add .\.agents\plugins
Get-Content plugins\codex-osf\.codex-plugin\plugin.json | ConvertFrom-Json
Get-Content plugins\codex-osf\.mcp.json | ConvertFrom-Json
```

Gemini CLI:

```powershell
gemini extensions install https://github.com/edithatogo/osf-cli-go --consent
Copy-Item -Recurse plugins\gemini-osf $env:USERPROFILE\.gemini\extensions\osf-cli-go
Get-Content $env:USERPROFILE\.gemini\extensions\osf-cli-go\gemini-extension.json | ConvertFrom-Json
```

Qwen Code:

```powershell
qwen extensions install https://github.com/edithatogo/osf-cli-go
Copy-Item -Recurse plugins\qwen-osf $env:USERPROFILE\.qwen\extensions\osf-cli-go
Get-Content $env:USERPROFILE\.qwen\extensions\osf-cli-go\qwen-extension.json | ConvertFrom-Json
```

Build self-contained plugin archives:

```powershell
.\scripts\build-plugin-archives.ps1
```

The `Plugin Archives` GitHub Actions workflow builds platform-specific ZIPs for
manual marketplace review or release attachment.

This repository publishes a GitHub-hosted Copilot marketplace at
`.github/plugin/marketplace.json`. That makes the package installable from the
public repository. It is not a claim that the package has been accepted into a
GitHub-maintained default marketplace; that provider review or organization
administration gate remains separately recorded.

Gemini CLI gallery discovery uses the root `gemini-extension.json` and the
`gemini-cli-extension` GitHub topic. The packaged `plugins/gemini-osf` variant
is used for self-contained release archives with a bundled `bin/osf-mcp`.

Qwen Code supports direct Git repository, local path, archive, npm, Claude
marketplace, and Gemini gallery installation. The root `qwen-extension.json`
supports direct repository installation, while `plugins/qwen-osf` is the
self-contained release-archive variant.
