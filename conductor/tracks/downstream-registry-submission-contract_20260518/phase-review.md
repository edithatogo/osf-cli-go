# Phase Review: Downstream Registry Submission Contract

Date: 2026-05-18

## Implemented

- Added `contract.md` with explicit Published/Prepared/Blocked terminology.
- Added `registry/directory-submissions.json` as the structured downstream
  directory contract tied to `server.json`.
- Added `tools/checkregistries` and wired it into local and CI checks.
- Added `registry/downstream-submission-packet.md` with manual form text and
  install/package fields for MCP.Directory, Glama, PulseMCP, Smithery, and
  Claude directory submission.
- Added MCPB package metadata in `packaging/mcpb/manifest.json`.
- Added `scripts/build-mcpb.ps1`; local run produced
  `dist/mcpb/osf-cli-go-0.2.0-win-x64.mcpb`.
- Added `.github/workflows/mcpb-bundles.yml` for platform MCPB artifacts.
- Added `scripts/build-plugin-archives.ps1`; local run produced Windows ZIPs
  for Claude, Codex, Gemini, and Qwen packages with bundled `osf-mcp.exe`.
- Added `.github/workflows/plugin-archives.yml` for platform plugin archives.
- Added `.mcp.json` and `plugins/github-copilot-osf` for GitHub Copilot
  repository/workspace MCP configuration packaging.
- Normalized Codex marketplace metadata and added sensitive OSF settings/env
  declarations to Gemini and Qwen extension manifests.
- Tightened `scripts/build-plugin-archives.ps1` so a failed `go build` stops
  the archive process instead of emitting partial ZIPs.

## Submission Status

- Smithery: published by CLI after browser authentication. Release
  `4a285e7c-567f-4c53-ae1d-af64e95fc054` was accepted for
  `edithatogo/osf-cli-go`; status `SUCCESS`; MCP URL
  `https://osf-cli-go--edithatogo.run.tools`.
- MCP.Directory: submitted by Playwright on 2026-05-18. The page returned
  `Server Submitted!`; local evidence is in
  `dist/submission-evidence/mcp-directory/result.json`.
- Glama: checked by Playwright. Search did not show `osf-cli-go`; the
  unauthenticated Add Server route did not expose a public submission form.
- PulseMCP: blocked. Browser and HTTP access to
  `https://www.pulsemcp.com/submit` returned Access Denied and asks submitters
  to contact `hello@pulsemcp.com`.
- MCP Central: checked by Playwright. Search did not show `osf-cli-go`; appears
  to be an aggregator/watch target.
- Claude plugin directory: prepared, requires external submission form/review.
- GitHub Copilot/Claude Cowork/Codex/Gemini/Qwen: local/release package routes
  prepared; public gallery or marketplace discovery depends on external
  install/review flows.
- Codex plugin marketplace: registered locally with
  `codex plugin marketplace add C:\Users\60217257\repos\osf-cli-go\.agents\plugins`.

## Validation Evidence

- `go run ./tools/checkregistries`: pass
- GitHub Copilot MCP JSON parse for `.github/mcp.json`, `.mcp.json`, and
  `plugins/github-copilot-osf`: pass
- JSON parse for `registry/directory-submissions.json`,
  `packaging/mcpb/manifest.json`, Codex marketplace, Gemini, and Qwen manifests:
  pass
- `scripts/build-mcpb.ps1`: pass locally, produced Windows `.mcpb`
- Smithery MCPB publish: pass after adding explicit MCP `inputSchema` objects
  to the six manifest tool entries.
- `scripts/build-plugin-archives.ps1`: pass locally with repo-local Go caches,
  produced five Windows ZIPs including `github-copilot-osf`
- `MCPB Bundles` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031259828`)
- `Plugin Archives` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031263119`)

Generated `dist/` output was removed after validation and is intentionally not
committed.
