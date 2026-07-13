# Metacheck OSF validation comparison

Reviewed 2026-07-13 against the public `main` branch of
[scienceverse/metacheck](https://github.com/scienceverse/metacheck).

## Evidence

The repository describes Metacheck as an experimental R package for checking
research outputs for best practices. At review time its public GitHub metadata
reported 45 stars, 11 forks, an AGPL-3.0 license, and a latest push on
2026-07-07. Its README includes Codecov and Zenodo badges and requires every
contribution to include tests.

Relevant source areas are the [OSF retrieval helpers](https://github.com/scienceverse/metacheck/blob/main/R/archive-osf.R),
[OSF helper functions](https://github.com/scienceverse/metacheck/tree/main/R),
[validation modules](https://github.com/scienceverse/metacheck/tree/main/inst/modules),
and [test fixtures](https://github.com/scienceverse/metacheck/tree/main/tests/testthat).
The package retrieves OSF node, child, and file metadata, identifies OSF links,
and contains modular checks for code, conflicts of interest, funding,
preregistration, repositories, and statistics. Its public documentation also
shows network and LLM-dependent checks are skipped or mocked in routine tests.

## Capability comparison

| Capability | Metacheck | OSF CLI Go | Decision |
|---|---|---|---|
| OSF identification and retrieval | R helpers accept IDs/URLs, fetch node metadata, and optionally recurse through children/files | Typed OSF API client, URL normalization, pagination, file traversal, and export | Existing client parity confirmed |
| Research-output metadata checks | Modular R checks and paper-oriented workflows | `validate --profile research-output` checks title, description, contributors, and storage presence | Implemented as a deterministic metadata slice |
| Preregistration checks | R module and paper-context checks | `validate --profile preregistration` checks title, description, contributors, and registration category | Implemented as a conservative metadata slice |
| Findings automation | R module output and interactive/reporting workflows | Stable JSON findings with rule, status, severity, message, profile, and validity | Implemented for CLI automation |
| Paper text, citations, statistics, and LLM checks | Domain-specific R modules, external services, and paper inputs | No paper parser, statistical interpreter, or LLM dependency | Rejected as outside a focused OSF client contract |
| OSF writes and credentials | Optional OSF PAT and network-dependent workflows | Explicit authenticated CLI writes with redaction and conservative confirmation | Existing security boundary retained |
| MCP exposure | No MCP server surface identified | Validation is CLI-only; MCP exposure is deferred with the existing read-only tool boundary | Deferred to MCP validation roadmap |
| Maturity and distribution | Experimental R package, Codecov, Zenodo, testthat, pkgdown | Go CLI/MCP, cross-platform release gates, race/lint/vulnerability checks | Different product surfaces; no wholesale port justified |

## Outcome

The accepted gap is a stable, read-only OSF metadata validation contract that
automation can consume without paper-text or LLM dependencies. The command
does not claim scientific validity and never modifies OSF. Metacheck's AGPL R
implementation, paper-specific modules, external-service behavior, and MCP
surface were not copied; they remain explicit deferred or rejected scope.
