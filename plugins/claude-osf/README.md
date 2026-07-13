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

## Install And Validate

```powershell
Copy-Item -Recurse plugins\claude-osf $env:USERPROFILE\.claude\plugins\osf-cli-go
claude plugin validate $env:USERPROFILE\.claude\plugins\osf-cli-go
```

For release packages, copy or unpack the archive that includes `bin\osf-mcp`
or `bin\osf-mcp.exe`; the local development command above is only for source
checkout testing.

## Official Directory Submission

Anthropic accepts public GitHub repository links or plugin ZIP uploads through
the [Claude.ai plugin submission form](https://claude.ai/settings/plugins/submit)
or the [Console plugin submission form](https://platform.claude.com/plugins/submit).
The repository source and the generated archive both contain the manifest,
MCP configuration, documentation, and bundled binary required for review.
Submission or Anthropic verification is not implied by local validation.
