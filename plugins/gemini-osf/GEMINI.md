# OSF CLI Go

Use the `osf` MCP server for authenticated, read-only Open Science Framework
inspection: current user, projects, components, files, and contributors.

Set `OSF_TOKEN` in the shell before starting Gemini CLI.

## Install And Validate

```powershell
Copy-Item -Recurse plugins\gemini-osf $env:USERPROFILE\.gemini\extensions\osf-cli-go
Get-Content $env:USERPROFILE\.gemini\extensions\osf-cli-go\gemini-extension.json | ConvertFrom-Json
```

For release packages, install from a generated
`gemini-osf-<version>-<runtime>.zip` archive so the MCP server binary is bundled
under `bin\`.
