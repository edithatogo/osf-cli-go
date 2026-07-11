# Downstream Submission Packet

Date: 2026-05-18

## Canonical Links

- Repository: `https://github.com/edithatogo/osf-cli-go`
- Official MCP Registry name: `io.github.edithatogo/osf-cli-go`
- Official MCP Registry lookup: `https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.edithatogo%2Fosf-cli-go`
- Official package: `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.3.0`
- Go module: `github.com/edithatogo/osf-cli-go@v0.3.0`
- Privacy policy for OSF service data: `https://osf.io/privacy-policy`
- Support/issues: `https://github.com/edithatogo/osf-cli-go/issues`

## Short Description

Read-only MCP tools for authenticated Open Science Framework projects and files.

## Long Description

OSF CLI Go provides a stdio Model Context Protocol server for inspecting an
authenticated Open Science Framework account. The server exposes read-only tools
for current-user identity, projects, components, OSF Storage files/folders, and
contributors. Authentication uses `OSF_TOKEN` by preference, with optional
`OSF_USERNAME` and `OSF_PASSWORD` fallback.

## Install Commands

Go:

```powershell
go install github.com/edithatogo/osf-cli-go/cmd/osf-mcp@v0.3.0
```

Docker/OCI:

```powershell
docker run --rm -i -e OSF_TOKEN ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.3.0
```

MCPB:

```powershell
.\scripts\build-mcpb.ps1 -UseMcpbCli
```

Plugin/extension ZIPs:

```powershell
.\scripts\build-plugin-archives.ps1
```

## Tools

- `osf_whoami`
- `osf_projects_list`
- `osf_project_get`
- `osf_components_list`
- `osf_files_list`
- `osf_contributors_list`

## Manual Submission Targets

Structured fields for directory forms are tracked in
`registry/directory-submissions.json`.

### MCP.Directory

Submit repository URL:

```text
https://github.com/edithatogo/osf-cli-go
```

Use the short and long descriptions above. Include official MCP Registry name
`io.github.edithatogo/osf-cli-go` if the form allows notes.

### Glama

Submit the GitHub repository and official registry name. If asked for the
runtime, select stdio MCP server and reference the OCI package.

### PulseMCP

Prefer official registry ingestion. If a manual form is used, submit repository
URL plus the official MCP Registry name.

### Smithery

Published with the MCPB route:

```powershell
smithery mcp publish .\dist\mcpb\osf-cli-go-0.3.0-<runtime>.mcpb -n edithatogo/osf-cli-go
```

Receipt:

```text
deploymentId: 4a285e7c-567f-4c53-ae1d-af64e95fc054
qualifiedName: edithatogo/osf-cli-go
status: SUCCESS
mcpUrl: https://osf-cli-go--edithatogo.run.tools
statusUrl: https://smithery.ai/servers/edithatogo/osf-cli-go/releases
```

### Claude Plugin Directory

Submit the public repository or ZIP containing `plugins/claude-osf`. Include
that the plugin uses the bundled `bin/osf-mcp` binary and read-only OSF tools.

### GitHub Copilot

Use `.github/mcp.json` for Copilot coding agent repository configuration and
`.mcp.json` for Copilot CLI/workspace MCP configuration. The release ZIP
`github-copilot-osf-<version>-<runtime>.zip` carries both files plus the bundled
`bin/osf-mcp` binary for review or attachment.
