# Project Tracks

This file tracks all major work items. Each track has its own spec and plan.
The per-track `plan.md` files are the source of truth for completion state.

---

- [x] [mvp-osf-readonly-cli](tracks/mvp-osf-readonly-cli_20260502/spec.md) — MVP read-only OSF CLI foundation.
- [x] [cli-contract-routing](tracks/cli-contract-routing_20260502/spec.md) — Cobra command contract, global flags, exit codes, help, and output behavior.
- [x] [api-client-fixtures](tracks/api-client-fixtures_20260502/spec.md) — OSF API v2 client, JSON:API pagination, typed errors, and fixture contracts.
- [x] [auth-public-access](tracks/auth-public-access_20260502/spec.md) — Token handling, public unauthenticated behavior, redaction, and `auth whoami`.
- [x] [readonly-commands](tracks/readonly-commands_20260502/spec.md) — Project/component/file read-only commands.
- [x] [download-safety](tracks/download-safety_20260502/spec.md) — Safe streaming downloads, conflict policies, manifests, and path protection.
- [x] [docs-release-readiness](tracks/docs-release-readiness_20260502/spec.md) — User docs, integration-test docs, release checklist, and examples.
- [x] [repo-quality-automation](tracks/repo-quality-automation_20260502/spec.md) — CI, linting, formatting, coverage, Renovate, and repo management.
- [x] [quality-review-automation](tracks/quality-review-automation_20260502/spec.md) — Anti-stub policy, phase review artifacts, and review-fix-continue workflow.
- [x] [mcp-server-roadmap](tracks/mcp-server-roadmap_20260502/spec.md) — CLI-first package boundaries for future MCP support.
- [x] [files-download-cli](tracks/files-download-cli_20260502/spec.md) — User-facing `files download` command backed by the safe download package.
- [x] [live-osf-validation](tracks/live-osf-validation_20260502/spec.md) — Opt-in live OSF validation scripts and evidence capture.
- [x] [release-packaging](tracks/release-packaging_20260502/spec.md) — Buildable release artifacts, version injection, completions, and release checks.
- [x] [coverage-hardening](tracks/coverage-hardening_20260502/spec.md) — Focused coverage improvements for auth, API, CLI error paths, and download safety.
- [x] [mcp-boundary-prep](tracks/mcp-boundary-prep_20260502/spec.md) — Minimal reusable boundaries for future MCP server development while keeping CLI-first delivery.
- [x] [osf-api-coverage](tracks/osf-api-coverage_20260502/spec.md) — Extended OSF API coverage to 20+ endpoints including registrations, wikis, comments, logs, identifiers, write operations.
- [x] [sota-repo-hardening](tracks/sota-repo-hardening_20260502/spec.md) — SOTA repo setup: issue templates, PR template, pre-commit hooks, dependabot, topics, license metadata.
- [x] [docs-overhaul](tracks/docs-overhaul_20260502/spec.md) — Comprehensive Go doc comments, install/usage/architecture docs, and documentation CI integration.
- [x] [waterbutler-write-operations](tracks/waterbutler-write-operations_20260502/spec.md) — File upload, folder creation, and file deletion via WaterButler API.
- [x] [preprints-search-registrations](tracks/preprints-search-registrations_20260502/spec.md) — Preprint listing, search, and add-on listing endpoints with CLI commands.
- [x] [node-export-commands](tracks/node-export-commands_20260502/spec.md) — `osf export` command with full node snapshot.
- [x] [username-password-auth](tracks/username-password-auth_20260517/spec.md) — Username/password authentication fallback with command-level capability mapping.
- [x] [conductor-state-reconciliation](tracks/conductor-state-reconciliation_20260517/spec.md) — Reconcile completed track index, plans, and stale phase-review status language.
- [x] [mcp-registry-plugin-distribution](tracks/mcp-registry-plugin-distribution_20260518/spec.md) — Implement MCP package/server distribution, registry submissions, and client plugins for Copilot, Claude/Cowork, Codex, Gemini CLI, and Qwen Code.
- [x] [downstream-registry-submission-contract](archive/downstream-registry-submission-contract_20260518/spec.md) — Complete Smithery/MCPB, downstream directory, and client-plugin submission contracts after official MCP Registry publication.
- [x] [official-mcp-github-registry-adoption](archive/official-mcp-github-registry-adoption_20260706/spec.md) — Maintain and improve Official MCP Registry and GitHub MCP Registry adoption for the OSF MCP server.
- [x] [smithery-quality-publication-adoption](archive/smithery-quality-publication-adoption_20260706/spec.md) — Optimize Smithery publication quality and adoption for the OSF MCP server.
- [x] [mcp-directory-adoption](archive/mcp-directory-adoption_20260706/spec.md) — Submit, verify, and optimize MCP.Directory adoption for the OSF MCP server.
- [x] [glama-quality-claim-adoption](archive/glama-quality-claim-adoption_20260706/spec.md) — Claim or improve Glama listing quality for the OSF MCP server.
- [ ] [cursor-directory-adoption](tracks/cursor-directory-adoption_20260706/spec.md) — Submit and optimize Cursor Directory adoption for the OSF MCP server.
- [ ] [claude-official-plugin-directory-adoption](tracks/claude-official-plugin-directory-adoption_20260706/spec.md) — Prepare and submit the OSF Claude plugin to the official Claude plugin directory.
- [ ] [codex-marketplace-adoption](tracks/codex-marketplace-adoption_20260706/spec.md) — Submit and optimize the OSF Codex plugin for Codex Marketplace adoption.
- [ ] [homebrew-tap-adoption](tracks/homebrew-tap-adoption_20260706/spec.md) — Prepare and submit Homebrew tap distribution for the OSF CLI and MCP binaries.
- [ ] [winget-adoption](tracks/winget-adoption_20260706/spec.md) — Prepare and submit WinGet distribution for OSF CLI Go.
- [ ] [scoop-adoption](tracks/scoop-adoption_20260706/spec.md) — Prepare and submit Scoop distribution for OSF CLI Go.

- [ ] [sourceshift-osf-mcp-parity](tracks/sourceshift-osf-mcp-parity_20260711/spec.md) — Benchmark SourceShift OSF MCP and close MCP capability gaps; source SourceShift/osf-mcp-server; issue #8.

- [ ] [osfclient-cli-parity](tracks/osfclient-cli-parity_20260711/spec.md) — Benchmark osfclient CLI workflows and close usability gaps; source osfclient/osfclient; issue #9.

- [ ] [osfr-entity-parity](tracks/osfr-entity-parity_20260711/spec.md) — Benchmark osfr entity coverage and conflict semantics; source ropensci/osfr; issue #10.

- [ ] [pyosf-sync-parity](tracks/pyosf-sync-parity_20260711/spec.md) — Benchmark pyosf synchronization and project workflows; source psychopy/pyosf; issue #11.

- [ ] [datalad-osf-parity](tracks/datalad-osf-parity_20260711/spec.md) — Benchmark DataLad OSF dataset and annex workflows; source datalad/datalad-osf; issue #12.

- [ ] [osf-sync-parity](tracks/osf-sync-parity_20260711/spec.md) — Benchmark OSF Sync desktop synchronization semantics; source CenterForOpenScience/osf-sync; issue #13.

- [ ] [datahugger-doi-parity](tracks/datahugger-doi-parity_20260711/spec.md) — Add DOI-oriented OSF retrieval parity with datahugger; source J535D165/datahugger; issue #14.

- [ ] [tooluniverse-osf-parity](tracks/tooluniverse-osf-parity_20260711/spec.md) — Add ToolUniverse-compatible OSF preprint and agent tooling; source mims-harvard/ToolUniverse; issue #15.

- [ ] [research-plugins-osf-parity](tracks/research-plugins-osf-parity_20260711/spec.md) — Add research-agent OSF discovery and full-text workflow parity; source wentorai/research-plugins; issue #16.

- [ ] [colrev-osf-parity](tracks/colrev-osf-parity_20260711/spec.md) — Add OSF literature-review ingestion and metadata parity; source CoLRev-Environment/colrev; issue #17.

- [ ] [jasp-osf-integration-parity](tracks/jasp-osf-integration-parity_20260711/spec.md) — Benchmark JASP OSF desktop integration workflows; source jasp-stats/jasp-desktop; issue #18.

- [ ] [presqt-osf-parity](tracks/presqt-osf-parity_20260711/spec.md) — Benchmark PresQT preservation and metadata-transfer workflows; source Lucy-Family-Institute/presqt; issue #19.

- [ ] [metacheck-osf-validation](tracks/metacheck-osf-validation_20260711/spec.md) — Add OSF research-output and preregistration validation workflows; source scienceverse/metacheck; issue #20.



- [x] [gemini-extension-gallery-publication](tracks/gemini-extension-gallery-publication_20260711/spec.md) — Publish the OSF extension to the Gemini CLI gallery; issue #23.

- [x] [qwen-extension-publication](tracks/qwen-extension-publication_20260711/spec.md) — Publish the OSF extension for Qwen Code; issue #24.

- [x] [coding-agent-ecosystem-publication](tracks/coding-agent-ecosystem-publication_20260711/spec.md) — Package and publish OSF integrations for additional coding agents; issue #25.

- [x] [mcp-catalog-discoverability-sweep](tracks/mcp-catalog-discoverability-sweep_20260711/spec.md) — Complete a high-value MCP registry and catalog sweep; issue #26.

- [x] [software-preprint-readiness](tracks/software-preprint-readiness_20260711/spec.md) — Prepare and assess an OSF CLI Go software preprint; issue #27.
