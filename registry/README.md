# Registry and Client Distribution

This directory records the package and submission surfaces for the OSF MCP
server.

## Official MCP Registry

Primary package route: OCI image on GitHub Container Registry.

Required local files:
- `server.json`
- `Dockerfile.mcp`

Publish flow:

```powershell
docker build -f Dockerfile.mcp -t ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.3.2 .
docker push ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.3.2
mcp-publisher login github
mcp-publisher publish
```

CI publish flow:

```powershell
git tag v0.3.2
git push origin v0.3.2
```

The `.github/workflows/mcp-registry.yml` workflow builds and pushes the GHCR
image, validates `server.json`, and publishes to the official MCP Registry with
GitHub OIDC.

Published official MCP Registry entry:
- `io.github.edithatogo/osf-cli-go`
- version `0.3.2`
- package `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.3.2`
- workflow run `https://github.com/edithatogo/osf-cli-go/actions/runs/26027015142`
- Go module tag `v0.3.2`

The image label `io.modelcontextprotocol.server.name` must match
`server.json` name: `io.github.edithatogo/osf-cli-go`.

## Smithery

Smithery can publish either a public Streamable HTTP endpoint or an MCPB bundle.
This repo uses the MCPB route for the stdio MCP server; a public Streamable
HTTP service is not required for the current distribution contract.

Published Smithery release:
- qualified name `edithatogo/osf-cli-go`
- deployment `4a285e7c-567f-4c53-ae1d-af64e95fc054`
- MCP URL `https://osf-cli-go--edithatogo.run.tools`
- status URL `https://smithery.ai/servers/edithatogo/osf-cli-go/releases`

The MCPB manifest is validated by `go run ./tools/checkregistries`, including
tool schema names and sensitive OSF auth configuration.

## MCP.Directory, Glama, PulseMCP

These directory submissions should point to the public GitHub repository after
the MCP server, install instructions, and at least one package route are
available in the default branch.

## Client Plugins

Client-specific install examples live under `plugins/` and `.github/mcp.json`.
All local development configs run `go run ./cmd/osf-mcp`; release bundles should
replace that with the packaged `osf-mcp` binary.
