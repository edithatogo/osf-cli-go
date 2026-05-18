# Spec: MCP Registry And Plugin Distribution

## Objective

Turn `osf-cli-go` from a CLI-only project with MCP roadmap notes into a
publishable agent-tooling surface: an MCP server/package where required,
registry-ready metadata, and installable plugins or extension manifests for the
agent clients the project should support.

## Target Registries And Surfaces

- Official MCP Registry / GitHub-hosted registry ecosystem.
- Smithery.
- MCP.Directory and other credible MCP discovery directories identified during
  research.
- Go module distribution through tags and the Go proxy/pkg.go.dev.
- GitHub Releases with binary artifacts and checksums.
- GitHub Copilot MCP configuration and/or extension path.
- Claude Code and Claude Cowork plugin packaging and submission.
- Codex plugin packaging.
- Gemini CLI extension packaging.
- Qwen Code extension compatibility and native metadata where needed.

## Acceptance Criteria

- Registry research is recorded with source URLs, submission prerequisites, and
  whether each target accepts CLI tools, MCP servers, hosted endpoints, MCPB
  bundles, package-registry artifacts, or plugins.
- The repo contains an actual MCP server or equivalent supported package
  artifact when a registry requires one; roadmap-only documentation is not
  treated as publishable implementation.
- Submission metadata is generated and validated for every selected target.
- External publication steps are executed where credentials, package ownership,
  and irreversible side effects are approved and available.
- Blocked submission steps are recorded with the exact blocker and next command
  or web action.
- Plugins/extensions are prepared for GitHub Copilot, Claude Code/Cowork,
  Codex, Gemini CLI, and Qwen Code, with validation instructions for each.
- Release/tag/package steps are performed only after the release checklist and
  safety review pass.

## Non-Goals

- Do not claim registry submission where the registry requires an MCP server and
  only the CLI exists.
- Do not publish live packages, tags, or registry entries without an auditable
  approval point and validated metadata.
- Do not store OSF credentials, GitHub tokens, or registry credentials in the
  repository.
