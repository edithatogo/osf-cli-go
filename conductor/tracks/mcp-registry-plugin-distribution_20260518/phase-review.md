# Phase Review: MCP Registry And Plugin Distribution

Date: 2026-05-18

## Implemented

- Added a real stdio MCP server entrypoint at `cmd/osf-mcp`.
- Added `internal/mcpserver` with read-only tools:
  - `osf_whoami`
  - `osf_projects_list`
  - `osf_project_get`
  - `osf_components_list`
  - `osf_files_list`
  - `osf_contributors_list`
- Added targeted MCP tests covering tool exposure, OSF URL normalization, storage path splitting, and path traversal rejection.
- Added Official MCP Registry metadata in `server.json` targeting GHCR OCI packaging.
- Added `Dockerfile.mcp` with the required `io.modelcontextprotocol.server.name` label.
- Updated GoReleaser to build both `osf` and `osf-mcp`.
- Added `.github/workflows/mcp-registry.yml` to build/push the GHCR image and
  publish `server.json` with GitHub OIDC.
- Added client/plugin surfaces:
  - GitHub Copilot repo MCP config: `.github/mcp.json`
  - GitHub Copilot coding-agent config template: `registry/github-copilot-coding-agent-mcp.json`
  - VS Code/Copilot config: `.vscode/mcp.json`
  - Claude Code/Cowork plugin: `plugins/claude-osf`
  - Codex plugin: `plugins/codex-osf`
  - Gemini CLI extension: `plugins/gemini-osf`
  - Qwen Code extension: `plugins/qwen-osf`

## Submission Status

Official MCP Registry submission is complete. Smithery, MCP.Directory, Glama,
and PulseMCP still require either a manual web submission, downstream indexing,
an MCPB artifact, or a public Streamable HTTP endpoint.

The prepared local route is:

1. Build and push `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`.
2. Run `mcp-publisher login github`.
3. Run `mcp-publisher publish` from the repository root.

The prepared CI route is:

1. Push the `v0.2.0` tag or manually dispatch `.github/workflows/mcp-registry.yml`.
2. Let GitHub Actions build/push the OCI image and publish `server.json` using
   GitHub OIDC.

The CI route was manually dispatched from `master` and completed successfully:

- Workflow run: `https://github.com/edithatogo/osf-cli-go/actions/runs/26027015142`
- GHCR image: `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`
- Registry name: `io.github.edithatogo/osf-cli-go`
- Registry version: `0.2.0`
- Registry status: `active`
- Published at: `2026-05-18T10:09:19.854378Z`
- Release tag: `v0.2.0`
- Go module proxy check: `go list -m -json github.com/edithatogo/osf-cli-go@v0.2.0`
  returned tag hash `49d0dc49cf40bc6bc85bdcdc19bbeea043cb45f1`.

The later tag-triggered registry workflow run rebuilt and pushed the same image,
then failed at the final publish step because the registry rejected duplicate
version `0.2.0`, as expected. The workflow was updated to treat that duplicate
publish case as idempotent success for future reruns.

Smithery remains blocked until either an MCPB artifact or a public Streamable
HTTP endpoint is produced.

## Validation

- `go test ./...`: pass
- `go vet ./...`: pass
- `go run ./tools/checkstubs`: pass
- `go run ./tools/checkreviews`: pass
- `go build ./cmd/osf ./cmd/osf-mcp`: pass
- `mkdocs build --strict`: pass
- `golangci-lint run`: pass, 0 issues
- `mcp-publisher validate`: pass against `https://registry.modelcontextprotocol.io`
- JSON manifest parse check: pass

Local Docker image build was attempted but blocked because the Docker daemon was
not running on this machine. `go run github.com/goreleaser/goreleaser/v2@latest
check` was attempted, but the toolchain compile/download did not complete in a
reasonable time and its Go processes were stopped.
