# Plan: MCP Registry And Plugin Distribution

## Phase 1: Registry And Plugin Surface Research

- [x] Task: Build a registry/submission matrix covering official MCP Registry,
  GitHub MCP registry surfaces, Smithery, MCP.Directory, Go/pkg.go.dev, GitHub
  Releases, and other credible MCP directories found during research.
- [x] Task: Build a plugin/extension matrix for GitHub Copilot, Claude
  Code/Cowork, Codex, Gemini CLI, and Qwen Code.
- [x] Task: Classify each target as immediate, requires MCP server, requires
  hosted endpoint, requires package artifact, requires manual web submission,
  or not applicable.

## Phase 2: MCP Server And Package Artifact

- [x] Task: Implement an MCP-facing OSF service layer without exposing Cobra or
  terminal formatting internals.
- [x] Task: Add a first read-only MCP tool set: whoami, projects list/get,
  components list, files list, and contributors list.
- [x] Task: Add stdio server packaging for local clients.
- [x] Task: Add HTTP/streamable deployment path or static server card if needed
  for Smithery and other URL-based registries. Commit: 37a65ff
- [x] Task: Add MCPB or OCI packaging if selected by the registry matrix.
- [x] Task: Add tests for MCP tool schemas, auth handling, redaction, and
  failure cases. Commit: 37b8131

## Phase 3: Registry Metadata And Release Readiness

- [x] Task: Generate `server.json` or equivalent metadata for official MCP
  Registry submission.
- [x] Task: Generate Smithery metadata/config schema or server-card endpoint.
  Commit: 37a65ff
- [~] Task: Prepare MCP.Directory submission metadata.
- [x] Task: Update release automation so publishing can produce current binaries
  and checksums from a new tag.
- [x] Task: Trigger Go proxy/pkg.go.dev visibility for the current release tag.
- [ ] Task: Run release checklist and record evidence.

## Phase 4: Client Plugins And Extensions

- [x] Task: Create a GitHub Copilot MCP configuration/extension guide for the OSF
  MCP server.
- [x] Task: Create Claude Code/Cowork plugin package with `.claude-plugin`,
  skills/commands, and MCP server configuration.
- [x] Task: Create Codex plugin package with `.codex-plugin/plugin.json`,
  skills/commands, and MCP server configuration.
- [x] Task: Create Gemini CLI extension package with `gemini-extension.json`.
- [x] Task: Create or validate Qwen Code extension compatibility, using
  converted Claude/Gemini metadata where supported.
- [~] Task: Add validation instructions and install commands for every plugin
  surface.

## Phase 5: Submission And Publication

- [x] Task: Submit to official MCP Registry after package/server prerequisites
  are satisfied.
- [ ] Task: Submit/publish to Smithery after URL/MCPB prerequisites are
  satisfied.
- [ ] Task: Submit to MCP.Directory and any selected additional MCP directories.
- [ ] Task: Submit Claude plugin to the Claude plugin directory when validation
  passes and public GitHub/ZIP package is ready.
- [ ] Task: Publish or document install paths for GitHub Copilot, Codex, Gemini
  CLI, and Qwen Code.
- [~] Task: Record submission receipts, URLs, review queues, or blockers.

## Phase 6: Review

- [ ] Task: Run `go test ./...`, `go vet ./...`, `golangci-lint run`,
  `go run ./tools/checkstubs`, and `go run ./tools/checkreviews`.
- [ ] Task: Run `$conductor-review`, apply fixes, rerun review, and write phase
  review evidence.
