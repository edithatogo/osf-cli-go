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
- Normalized Codex marketplace metadata and added sensitive OSF settings/env
  declarations to Gemini and Qwen extension manifests.

## Submission Status

- Smithery: prepared for MCPB route, not submitted because Smithery CLI auth is
  not installed/logged in locally and the artifact should come from release/CI.
- MCP.Directory: prepared, manual form at `https://mcp.directory/submit`.
- Glama: prepared, expected official-registry ingestion or web/OAuth claim.
- PulseMCP: prepared, expected official-registry ingestion or manual submit at
  `https://www.pulsemcp.com/submit`.
- Claude plugin directory: prepared, requires external submission form/review.
- Codex/Gemini/Qwen: local/release package routes prepared; public gallery or
  marketplace discovery depends on external install/review flows.

## Validation Evidence

- `go run ./tools/checkregistries`: pass
- JSON parse for `registry/directory-submissions.json`,
  `packaging/mcpb/manifest.json`, Codex marketplace, Gemini, and Qwen manifests:
  pass
- `scripts/build-mcpb.ps1`: pass locally, produced Windows `.mcpb`
- `scripts/build-plugin-archives.ps1`: pass locally, produced four Windows ZIPs
- `MCPB Bundles` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031259828`)
- `Plugin Archives` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031263119`)

Generated `dist/` output was removed after validation and is intentionally not
committed.
