---
title: "osf-cli-go: A reproducible command-line and MCP interface for the Open Science Framework"
tags:
  - Open Science Framework
  - research software
  - command-line interface
  - Model Context Protocol
  - reproducibility
authors:
  - name: Edith Atogo
    affiliation: 1
affiliations:
  - name: Independent researcher
    index: 1
date: 2026-07-11
bibliography: references.bib
---

# Summary

`osf-cli-go` is an Apache-2.0 command-line client and stdio Model Context
Protocol (MCP) server for the Open Science Framework (OSF). It provides
machine-readable and human-readable workflows for authentication, project and
component inspection, file transfer, search, preprints, registrations, and
node export. The MCP surface exposes conservative read-oriented tools while
the CLI keeps explicit confirmation boundaries for write operations.

# Statement of Need

OSF hosts research projects, registrations, files, contributors, and preprints,
but automation is often split between web workflows, language-specific clients,
and ad hoc scripts. This makes reproducible setup, pagination, error handling,
credential management, and agent integration inconsistent. `osf-cli-go`
provides a single Go implementation with deterministic JSON output, fixture-
backed API behavior, safe download semantics, opt-in live validation, and
portable client configuration for MCP-capable coding agents.

# Design and Implementation

The implementation separates CLI routing, authentication, API transport,
output rendering, download safety, and MCP server plumbing. OSF JSON:API
pagination and typed errors are handled behind testable interfaces. Credentials
are read from environment variables, token values are redacted, and live OSF
checks are opt-in. File operations use explicit conflict policies, manifests,
path protection, and atomic writes. Destructive or state-changing operations
require explicit confirmation.

The repository publishes `server.json`, an OCI package contract, client
metadata for Codex, Claude, Copilot, Gemini, and Qwen, and standard MCP
configuration templates for Cursor, Cline, Roo Code, Windsurf, VS Code, and
Zed. These surfaces are documented as available or prepared; provider approval
is never inferred from local artifacts.

# Reproducibility and Quality

The repository includes fixture-backed unit tests, race tests, `go vet`,
anti-stub checks, registry checks, release-contract validation, CI workflows,
and reproducible package/archive builders. The release contract checks version
alignment across the server and client manifests, credential-safe integration
templates, and catalog metadata. Live OSF validation is intentionally separate
from ordinary CI because it requires credentials and a disposable project.

# Comparative Position

The project is evaluated against OSF ecosystem references including `osfclient`,
`osfr`, SourceShift's OSF MCP server, and export/synchronization tools. The
maintained comparison records both parity and deliberate scope boundaries.
`osf-cli-go` emphasizes a typed Go implementation, explicit safety semantics,
deterministic automation output, MCP distribution, and cross-client
configuration rather than claiming every ecosystem workflow is complete.

# Availability

Source code, tests, documentation, release metadata, and issue history are
available at <https://github.com/edithatogo/osf-cli-go>. The current release
candidate for this manuscript is `v0.3.1`. A DOI-backed archive and formal
venue submission are intentionally deferred pending human author review,
release-candidate gates, and stronger evidence of external research adoption.

# AI Usage Disclosure

Generative AI tools assisted with repository inspection, documentation drafting,
and validation orchestration. The author is responsible for verifying all
technical claims, references, evaluation results, authorship, and submission
decisions. No AI system is an author or reviewer, and no external manuscript or
preprint has been submitted by this track.

# References

See `references.bib` for the OSF API, MCP, and comparative software references.
