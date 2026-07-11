# MCP Catalog Discoverability Evidence

Last reviewed: 2026-07-11

## Authoritative references

- [Docker MCP Catalog contribution guide](https://docs.docker.com/ai/mcp-catalog-and-toolkit/catalog/)
- [Docker MCP Registry contributing guide](https://github.com/docker/mcp-registry/blob/main/CONTRIBUTING.md)
- [mcp.so](https://mcp.so/?tab=official)
- [MCP.Directory submission](https://mcp.directory/submit)
- [MCPize documentation](https://docs.mcpize.com/)
- [Awesome MCP Servers submission notice](https://github.com/wong2/awesome-mcp-servers)

## Prepared submission

Docker's maintained path is a PR to `docker/mcp-registry` with a
`servers/osf-cli-go/server.yaml` entry. The repository packet is stored at
`registry/docker-mcp-registry/` and includes the required server metadata and
tool inventory. Docker can build the image from this repository's
`Dockerfile.mcp`; no credentials are included.

The packet uses the public source commit `2047a2a`, the current pushed release
contract, and the three optional OSF credential variables as secrets.

## Target matrix

| Target | Route | Local status | External status |
| --- | --- | --- | --- |
| Docker MCP Catalog | PR to `docker/mcp-registry` | Submission packet ready; Dockerfile and tool inventory present | PR/review not claimed |
| mcp.so | GitHub issue submission | Public submission details prepared | Submission not claimed |
| awesome-mcp-servers / mcpservers.org | Website submission linked by repository | Public README and metadata ready | Submission not claimed |
| MCP.Directory | `/submit` form | Public repository URL and metadata ready | Existing pending state retained; new approval not claimed |
| PulseMCP | Official registry ingest or manual contact | Official registry metadata published | Access/submission gate remains unresolved |
| MCPize | GitHub App/account deployment | Docker and public repository assets ready | Account/deployment not claimed |

The project distinguishes prepared, submitted, under review, published, and
blocked states. No catalog is described as published without dated provider
evidence.

## Local validation

```text
go run ./tools/checkreleasecontract
go test ./...
go run ./tools/checkregistries
```

Docker image smoke testing and provider review are external gates because they
require Docker/registry state or third-party submission systems.
