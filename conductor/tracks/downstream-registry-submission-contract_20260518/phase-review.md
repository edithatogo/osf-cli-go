# Phase Review: Downstream Registry Submission Contract

Date: 2026-07-06

## Implemented

- Added `contract.md` with explicit Published/Prepared/Blocked terminology.
- Added `registry/directory-submissions.json` as the structured downstream
  directory contract tied to `server.json`.
- Added `tools/checkregistries` and wired it into local and CI checks.
- Added `registry/downstream-submission-packet.md` with manual form text and
  install/package fields for MCP.Directory, Glama, PulseMCP, Smithery, and
  Claude directory submission.
- Added MCPB package metadata in `packaging/mcpb/manifest.json`.
- Added `scripts/build-mcpb.ps1` and `.github/workflows/mcpb-bundles.yml` for
  platform MCPB artifacts.
- Added `scripts/build-plugin-archives.ps1` and
  `.github/workflows/plugin-archives.yml` for platform plugin archives.
- Added `.mcp.json` and `plugins/github-copilot-osf` for GitHub Copilot
  repository/workspace MCP configuration packaging.
- Normalized Codex marketplace metadata and added sensitive OSF settings/env
  declarations to Gemini and Qwen extension manifests.
- Tightened plugin/archive scripts so failed `go build` stops archive creation.
- Validated the MCPB manifest through `go run ./tools/checkregistries`,
  including tool schemas, sensitive OSF auth config, and Smithery evidence.

## Submission Status

- Smithery: Published. Release `4a285e7c-567f-4c53-ae1d-af64e95fc054` was
  accepted for `edithatogo/osf-cli-go`; status `SUCCESS`; MCP URL
  `https://osf-cli-go--edithatogo.run.tools`.
- MCP.Directory: Prepared. Browser submission returned `Server Submitted!`, but
  external listing/review remains outside repo-local control.
- Glama: Blocked. Search did not show `osf-cli-go`; the unauthenticated Add
  Server route did not expose a public submission form.
- PulseMCP: Blocked. Browser and HTTP access to
  `https://www.pulsemcp.com/submit` returned Access Denied and asks submitters
  to contact `hello@pulsemcp.com`.
- MCP Central: Prepared/watch target. It appears to be an aggregator rather
  than an independent submission target.
- Claude plugin directory: Prepared, requires external submission form/review.
- GitHub Copilot, Claude Cowork, Codex, Gemini, and Qwen: Prepared via
  repo-local/release package routes; public discovery depends on external
  install/review flows.

## Validation Evidence

- `go run ./tools/checkregistries`: pass
- JSON parse for `server.json`, `registry/directory-submissions.json`,
  `packaging/mcpb/manifest.json`, Claude, Codex, Gemini, and Qwen manifests:
  pass
- `go test ./...`: pass
- `go vet ./...`: pass
- `go run ./tools/checkstubs`: pass
- `go run ./tools/checkreviews`: pass

Previous external workflow evidence remains:

- `MCPB Bundles` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031259828`)
- `Plugin Archives` workflow: pass on Linux, macOS, and Windows
  (`https://github.com/edithatogo/osf-cli-go/actions/runs/26031263119`)

Generated `dist/` output was removed after validation and is intentionally not
committed.

## Conductor Review

- Review command: `$conductor-review downstream-registry-submission-contract_20260518`
- Blocking findings: none remaining
- Fixes applied: normalized MCP.Directory and Codex statuses back to the
  track's Published/Prepared/Blocked contract and refreshed stale phase-review
  evidence
- Re-review result: pass after evidence update
