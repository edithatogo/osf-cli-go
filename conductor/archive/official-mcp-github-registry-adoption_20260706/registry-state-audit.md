# Registry State Audit

Date: 2026-07-07

## Official MCP Registry

- Registry base: `https://registry.modelcontextprotocol.io`
- Server name: `io.github.edithatogo/osf-cli-go`
- Latest version: `0.2.0`
- Package: `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`
- Repository: `https://github.com/edithatogo/osf-cli-go`
- Registry search URL:
  `https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.edithatogo%2Fosf-cli-go`

Live API evidence:

```json
{
  "server": {
    "name": "io.github.edithatogo/osf-cli-go",
    "version": "0.2.0",
    "packages": [
      {
        "registryType": "oci",
        "identifier": "ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0"
      }
    ]
  },
  "_meta": {
    "io.modelcontextprotocol.registry/official": {
      "status": "active",
      "isLatest": true
    }
  }
}
```

## GitHub MCP Registry Ecosystem

- GitHub docs confirm MCP registry support for Copilot and enterprise
  configuration.
- The repo-local package metadata already uses the GitHub namespace path
  `io.github.edithatogo/osf-cli-go`, which matches the official registry entry.

## Validation

- `go test ./tools/checkregistries`
- `curl -fsS 'https://registry.modelcontextprotocol.io/v0.1/servers?search=osf-cli-go&limit=20'`
- `curl -fsS 'https://registry.modelcontextprotocol.io/v0.1/servers/io.github.edithatogo%2Fosf-cli-go/versions/latest'`

## Notes

- The official registry entry is live.
- The repo-local checker now verifies the official registry search URL, OCI
  package type, auth-secret flags, directory metadata coverage, and install
  snippet consistency.
