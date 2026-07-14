# Project Tracks

This file tracks all major work items. Each track has its own spec and plan.
The per-track `plan.md` files are the source of truth for completion state.

---

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
- [x] [codex-marketplace-adoption](archive/codex-marketplace-adoption_20260706/spec.md) — Submit and optimize the OSF Codex plugin for Codex Marketplace adoption.

- [x] [osfclient-cli-parity](tracks/osfclient-cli-parity_20260711/spec.md) — Benchmark osfclient CLI workflows and close usability gaps; source osfclient/osfclient; issue #9.

- [x] [osfr-entity-parity](tracks/osfr-entity-parity_20260711/spec.md) — Benchmark osfr entity coverage and conflict semantics; source ropensci/osfr; issue #10.

- [x] [pyosf-sync-parity](tracks/pyosf-sync-parity_20260711/spec.md) — Benchmark pyosf synchronization and project workflows; source psychopy/pyosf; issue #11.

- [x] [osf-sync-parity](tracks/osf-sync-parity_20260711/spec.md) — Benchmark OSF Sync desktop synchronization semantics; source CenterForOpenScience/osf-sync; issue #13.

- [x] [research-plugins-osf-parity](tracks/research-plugins-osf-parity_20260711/spec.md) — Add research-agent OSF discovery and full-text workflow parity; source wentorai/research-plugins; issue #16.


- [x] [presqt-osf-parity](tracks/presqt-osf-parity_20260711/spec.md) — Benchmark PresQT preservation and metadata-transfer workflows; source Lucy-Family-Institute/presqt; issue #19.




- [x] [gemini-extension-gallery-publication](tracks/gemini-extension-gallery-publication_20260711/spec.md) — Publish the OSF extension to the Gemini CLI gallery; issue #23.

- [x] [qwen-extension-publication](tracks/qwen-extension-publication_20260711/spec.md) — Publish the OSF extension for Qwen Code; issue #24.

- [x] [coding-agent-ecosystem-publication](tracks/coding-agent-ecosystem-publication_20260711/spec.md) — Package and publish OSF integrations for additional coding agents; issue #25.

- [x] [mcp-catalog-discoverability-sweep](tracks/mcp-catalog-discoverability-sweep_20260711/spec.md) — Complete a high-value MCP registry and catalog sweep; issue #26.

- [x] [software-preprint-readiness](tracks/software-preprint-readiness_20260711/spec.md) — Prepare and assess an OSF CLI Go software preprint; issue #27.

- [x] [paperclip-research-retrieval-parity](tracks/paperclip-research-retrieval-parity_20260713/spec.md) — Benchmark Paperclip research retrieval parity; issue #45.
- [~] [cline-mcp-marketplace-adoption](tracks/cline-mcp-marketplace-adoption_20260713/spec.md) — Submit OSF MCP server to the official Cline MCP Marketplace; issue #47.
- [x] [lobehub-mcp-marketplace-adoption](tracks/lobehub-mcp-marketplace-adoption_20260713/spec.md) — Submit OSF MCP server to LobeHub Marketplace; issue #48.
- [x] [additional-mcp-registry-sweep](tracks/additional-mcp-registry-sweep_20260713/spec.md) — Evaluate additional MCP registries and agent marketplaces; issue #49.
- [x] [mcp-quality-evaluation-harness](tracks/mcp-quality-evaluation-harness_20260713/spec.md) — Build repeatable MCP quality and compatibility evaluation; issue #54.
- [x] [osf-api-entity-coverage](tracks/osf-api-entity-coverage_20260714/spec.md) — Expose typed OSF file-version and node-related read surfaces through CLI and MCP; issue #80.
- [x] [datalad-git-annex-interoperability](tracks/datalad-git-annex-interoperability_20260714/spec.md) — Define a fixture-backed optional Git, DataLad, and git-annex companion-tool contract; issue #69.
- [x] [resumable-transfers](archive/resumable-transfers_20260714/spec.md) — Implement resumable, checkpointed transfers for 1.0; issue #95.
- [x] [structured-observability](archive/structured-observability_20260714/spec.md) — Define and implement structured observability for 1.0; issue #96.
- [~] [live-osf-release-validation](tracks/live-osf-release-validation_20260714/spec.md) — Run live OSF validation for the 1.0 release candidate; issue #97.
- [x] [release-candidate-provenance](tracks/release-candidate-provenance_20260714/spec.md) — Verify v1.0 release-candidate supply chain and provenance; issue #98.
- [x] [compatibility-contract-freeze](archive/compatibility-contract-freeze_20260714/spec.md) — Freeze OSF API and CLI/MCP compatibility contracts for 1.0; issue #99.
- [ ] [zenodo-provider-abstraction](tracks/zenodo-provider-abstraction_20260715/spec.md) — Add Zenodo support through a provider-neutral OSF/Zenodo research repository model; issue #101.
- [x] [zenodo-api-provenance](archive/zenodo-api-provenance_20260715/spec.md) — Pin Zenodo API evidence and detect upstream contract drift; subissue #102.
- [x] [repository-provider-contract](archive/repository-provider-contract_20260715/spec.md) — Define capability-aware, lossless provider contracts; subissue #103.
- [x] [zenodo-readonly-client](archive/zenodo-readonly-client_20260715/spec.md) — Implement an offline-tested read-only Zenodo REST client; subissue #104.
- [x] [provider-scoped-cli](archive/provider-scoped-cli_20260715/spec.md) — Add explicit provider-scoped CLI discovery workflows; subissue #105.
- [x] [provider-scoped-mcp](archive/provider-scoped-mcp_20260715/spec.md) — Expose capability-aware provider-scoped MCP read tools; subissue #106.
- [x] [zenodo-oai-pmh](archive/zenodo-oai-pmh_20260715/spec.md) — Implement Zenodo OAI-PMH harvesting as a separate adapter; subissue #107.
- [ ] [zenodo-sandbox-transfers](tracks/zenodo-sandbox-transfers_20260715/spec.md) — Validate safe resumable Zenodo transfers in the sandbox; subissue #108.
- [ ] [zenodo-publication-state](tracks/zenodo-publication-state_20260715/spec.md) — Model DOI and publication workflows as an irreversible state machine; subissue #109.
- [ ] [cross-provider-provenance-transfer](tracks/cross-provider-provenance-transfer_20260715/spec.md) — Design explicit cross-provider copy with provenance and recovery; subissue #110.
- [ ] [multi-provider-release-integration](tracks/multi-provider-release-integration_20260715/spec.md) — Integrate provider validation into CI, releases, observability, and docs; subissue #111.
