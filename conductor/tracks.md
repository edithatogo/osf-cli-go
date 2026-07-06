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
- [~] [downstream-registry-submission-contract](tracks/downstream-registry-submission-contract_20260518/spec.md) — Complete Smithery/MCPB, downstream directory, and client-plugin submission contracts after official MCP Registry publication.
