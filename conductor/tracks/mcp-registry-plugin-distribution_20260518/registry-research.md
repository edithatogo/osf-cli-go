# Registry And Plugin Research

Last updated: 2026-05-18.

## Current Repo Baseline

- `cmd/osf-mcp` now provides a real stdio MCP server backed by
  `internal/mcpserver` and the existing `internal/osfapi` client.
- The current GitHub release `v0.1.0` points to an older commit and has no
  release assets. The current `master` commit has not been tagged for release.
- `.goreleaser.yaml` now builds both `osf` and `osf-mcp`, but live release
  publishing remains disabled.

## Registry Matrix

| Target | Current fit | Submission path | Requirements / blockers |
| --- | --- | --- | --- |
| Official MCP Registry / modelcontextprotocol registry | Metadata prepared; live publish blocked until OCI image exists | `mcp-publisher publish` with `server.json` after login | `server.json` targets `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`; image must be built and pushed with matching `io.modelcontextprotocol.server.name` label first. |
| GitHub MCP registry ecosystem | Same as official MCP Registry plus GitHub Copilot configs | Publish via official MCP registry metadata and GitHub namespace/package ownership | GitHub namespace path is `io.github.edithatogo/osf-cli-go`; Copilot local configs are prepared in `.github/mcp.json` and `registry/github-copilot-coding-agent-mcp.json`. |
| Smithery | Blocked until HTTP MCP endpoint or MCPB bundle exists | Web flow at Smithery or `smithery mcp publish "https://..." -n @org/name` | URL publishing requires Streamable HTTP transport; local stdio requires pre-built MCPB bundle. |
| MCP.Directory | Ready after MCP implementation and install docs are on default branch | Submit GitHub repository URL via form | Auto-detects metadata/tools from GitHub; optional npm/PyPI package fields. |
| Go proxy / pkg.go.dev | Partially done for old `v0.1.0`; current code not published | Tag release and run `GOPROXY=proxy.golang.org go list -m github.com/edithatogo/osf-cli-go@vX.Y.Z` | Needs new semver tag for current code. |
| GitHub Releases | Old release exists with no assets | Create release/tag with binaries and checksums | GoReleaser now builds both binaries, but publishing is still disabled until release approval. |

## Plugin And Extension Matrix

| Target | Current fit | Packaging/submission path | Requirements / blockers |
| --- | --- | --- | --- |
| GitHub Copilot | Local/repo MCP config prepared | `.github/mcp.json`; coding-agent template under `registry/` | Coding agent secret names must use `COPILOT_MCP_` prefix. |
| Claude Code | Plugin package prepared | `.claude-plugin/plugin.json`, root `.mcp.json`; validate with `claude plugin validate`; distribute via marketplace or submit directory | Public repo or ZIP required for directory submission. |
| Claude Cowork | Claude plugin package can be distributed by admin/org plugin path | Cowork plugin package or Claude plugin directory submission | Bundled binary path must exist before org deployment. |
| Codex | Plugin package and repo marketplace metadata prepared | `.codex-plugin/plugin.json`, `.mcp.json`, skill, `.agents/plugins/marketplace.json` | Bundled binary path must exist for packaged install. |
| Gemini CLI | Native extension package prepared | `gemini-extension.json`; install examples use `gemini extensions install <github-url>` | Release archive must place manifest at archive root. |
| Qwen Code | Native extension package prepared | `qwen-extension.json` or converted Gemini extension | Release archive must place manifest at archive root. |

## Source Notes

- Official MCP Registry docs: https://modelcontextprotocol.io/registry/about
- MCP package types: https://modelcontextprotocol.io/registry/package-types
- MCP publish quickstart: https://modelcontextprotocol.io/registry/quickstart
- Smithery publish docs: https://smithery.ai/docs/build/publish
- MCP.Directory submit page: https://mcp.directory/submit
- Go module publishing docs: https://tip.golang.org/doc/modules/publishing
- Claude plugin submission docs: https://claude.com/docs/plugins/submit
- Claude plugin overview: https://claude.com/docs/plugins/overview
- Claude Cowork organization plugin docs: https://claude.com/docs/cowork/3p/extensions
- GitHub Copilot MCP setup docs: https://docs.github.com/en/copilot/how-tos/provide-context/use-mcp/set-up-the-github-mcp-server
- Gemini extension gallery examples: https://geminicli.com/extensions/
- Qwen Code extension compatibility docs: https://qwenlm.github.io/qwen-code-docs/en/users/extension/introduction/
