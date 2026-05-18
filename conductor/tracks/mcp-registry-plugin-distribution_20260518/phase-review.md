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

No live registry submission has been completed yet. The official MCP Registry,
Smithery, MCP.Directory, Glama, and PulseMCP all require a committed public
server plus at least one usable package, release artifact, hosted endpoint, or
manual web review step.

The prepared local route is:

1. Build and push `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`.
2. Run `mcp-publisher login github`.
3. Run `mcp-publisher publish` from the repository root.

The prepared CI route is:

1. Push the `v0.2.0` tag or manually dispatch `.github/workflows/mcp-registry.yml`.
2. Let GitHub Actions build/push the OCI image and publish `server.json` using
   GitHub OIDC.

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
