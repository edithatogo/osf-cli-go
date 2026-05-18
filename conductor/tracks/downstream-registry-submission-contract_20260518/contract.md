# Downstream Submission Contract

Date: 2026-05-18

## Status Language

- **Published** means a live endpoint accepted the package and a receipt or API
  response is recorded.
- **Prepared** means repo-local artifacts and instructions are ready, but a live
  submission or external review has not happened.
- **Blocked** means a target requires a hosted endpoint, manual web form,
  credential, signing key, or review process that is not available from this
  automated repo workflow.

## Target Matrix

| Target | Status | Required artifact | Submission path | Evidence required |
| --- | --- | --- | --- | --- |
| Official MCP Registry | Published | `server.json`, GHCR OCI image | GitHub OIDC workflow | Registry API response and workflow run |
| Smithery | Prepared | MCPB bundle or public Streamable HTTP endpoint | `smithery mcp publish ./artifact.mcpb -n edithatogo/osf-cli-go` or URL publish | Smithery publish output or blocker |
| MCP.Directory | Prepared | Public GitHub repo and MCP registry entry | Manual submit form | Submitted URL or manual blocker |
| Glama | Prepared | Public GitHub repo or official registry listing | Manual/listing flow | Submitted URL or indexing URL |
| PulseMCP | Prepared | Official registry listing or manual submit form | Registry ingestion or manual submit | Listing URL or manual blocker |
| Claude plugin directory | Prepared | Public repo/ZIP plugin package and validation output | Claude plugin submission form | Submission receipt or manual blocker |
| Claude Cowork org plugin | Prepared | `plugins/claude-osf` plus bundled binary | Admin/org deployment path | Deployment path receipt or blocker |
| Codex plugin | Prepared | `plugins/codex-osf` and `.agents/plugins/marketplace.json` | Repo/local plugin install | Install/validation output |
| Gemini CLI extension | Prepared | `plugins/gemini-osf` with bundled binary | GitHub/local extension install | Install/validation output |
| Qwen Code extension | Prepared | `plugins/qwen-osf` with bundled binary | GitHub/local extension install | Install/validation output |

## Submission Rules

- Prefer automated CLI/API workflows when credentials are already available.
- Treat manual web forms as prepared unless a browser submission receipt is
  captured.
- Treat duplicate registry publish errors as success only when the existing
  version is verified live and unchanged.
- Keep generated bundles and binaries under `dist/`; do not commit them.
- Never include `OSF_TOKEN`, `OSF_USERNAME`, or `OSF_PASSWORD` values in
  package metadata, screenshots, logs, or submission packets.
