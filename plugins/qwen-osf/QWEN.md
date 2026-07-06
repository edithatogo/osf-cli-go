# OSF CLI Go

Use the `osf` MCP server for authenticated, read-only Open Science Framework
inspection: current user, projects, components, files, and contributors.

Set `OSF_TOKEN` in the shell before starting Qwen Code.

## Install And Validate

```powershell
Copy-Item -Recurse plugins\qwen-osf $env:USERPROFILE\.qwen\extensions\osf-cli-go
Get-Content $env:USERPROFILE\.qwen\extensions\osf-cli-go\qwen-extension.json | ConvertFrom-Json
```

For release packages, install from a generated
`qwen-osf-<version>-<runtime>.zip` archive so the MCP server binary is bundled
under `bin\`.
