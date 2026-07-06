# Submission Closeout Evidence

Date: 2026-07-06

Scope: final submission/publication status for the MCP registry/plugin
distribution track.

## Published Or Submitted

| Target | Status | Evidence |
| --- | --- | --- |
| Official MCP Registry | Published | GitHub Actions run `https://github.com/edithatogo/osf-cli-go/actions/runs/26027015142`; package `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0` |
| Smithery | Published | Deployment `4a285e7c-567f-4c53-ae1d-af64e95fc054`; MCP URL `https://osf-cli-go--edithatogo.run.tools`; status URL `https://smithery.ai/servers/edithatogo/osf-cli-go/releases` |
| MCP.Directory | Submitted | Playwright submission returned `Server Submitted!`; evidence path `dist/submission-evidence/mcp-directory/result.json` |

## Prepared Install Paths

| Target | Repo-local artifact | Validation/install evidence |
| --- | --- | --- |
| GitHub Copilot | `.github/mcp.json`, `.mcp.json`, `plugins/github-copilot-osf/` | `plugins/README.md`, `plugins/github-copilot-osf/README.md`, plugin archive workflow |
| Claude Code/Cowork | `plugins/claude-osf/`, `.claude-plugin/marketplace.json` | `plugins/README.md`, `plugins/claude-osf/README.md`, plugin archive workflow |
| Codex | `plugins/codex-osf/`, `.agents/plugins/marketplace.json` | `plugins/README.md`, `plugins/codex-osf/skills/osf/SKILL.md` |
| Gemini CLI | `plugins/gemini-osf/` | `plugins/README.md`, `plugins/gemini-osf/GEMINI.md` |
| Qwen Code | `plugins/qwen-osf/` | `plugins/README.md`, `plugins/qwen-osf/QWEN.md` |

## External Gates

- Claude plugin directory: prepared for repository or ZIP submission, but final
  listing requires the provider's external submission/review flow.
- Claude Cowork/org deployment: prepared package route exists, but organization
  deployment requires admin-owned installation outside this repository.
- GitHub Copilot, Codex, Gemini CLI, and Qwen Code public gallery discovery:
  local/release package install paths are documented; public marketplace
  listing depends on provider-specific review or marketplace support.
- Glama: checked previously; unauthenticated add-server route did not expose a
  public submission form.
- PulseMCP: `https://www.pulsemcp.com/submit` returned Access Denied and asks
  submitters to contact `hello@pulsemcp.com`.

## Validation

- `go run ./tools/checkregistries`: pass on 2026-07-06
- Plugin JSON parse via `jq empty`: pass on 2026-07-06
- `go test ./...`: pass on 2026-07-06
- `go run ./tools/checkstubs`: pass on 2026-07-06

No new live registry submission was attempted in this pass. This evidence
records the already-published/submitted receipts and the remaining external
review gates.
