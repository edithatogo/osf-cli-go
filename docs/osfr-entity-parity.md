# osfr entity parity

Last reviewed: 2026-07-14

This comparison uses the public [ropensci/osfr repository](https://github.com/ropensci/osfr), its [README](https://github.com/ropensci/osfr/blob/master/README.md), and the current OSF CLI Go contracts.

## Source maturity

| Signal | osfr | OSF CLI Go |
|---|---|---|
| License | MIT | Apache-2.0 |
| Maintenance | 1,088 commits, 151 stars, 31 forks, 13 releases; latest shown as v0.2.9 from 2022-09-25 | Active Go CLI/MCP project with current release, security, registry, test, race, and vet gates |
| Runtime | R package distributed through CRAN or GitHub | Cross-platform Go binaries, JSON CLI, and stdio MCP server |
| Entity model | R `osf_tbl` abstractions for nodes, files/folders, and users | Typed Go API models for projects/nodes, components, contributors, files, preprints, and storage providers |
| Workflow | Public exploration, browser open, download, project/component creation, directory creation, and upload | Public reads, explicit downloads, authenticated writes, path/conflict controls, and structured output |
| Testing and automation | R package tests, R-CMD-check, Codecov, pkgdown, and CRAN distribution | Offline HTTP fixtures, deterministic CLI/MCP tests, anti-stub, race, vet, release, and security gates |

## Capability comparison

| Capability | osfr reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| Nodes and components | Retrieve nodes, list child nodes, create projects/components | Projects and components list/get/create/update/delete with normalized identifiers | Implemented |
| Files and folders | List files, download files, create directories, upload, open | List, explicit file/tree download, mkdir, upload, delete, open, conflict and path safety | Implemented with stronger transfer controls |
| Users and contributors | User/entity tables and contributor-oriented metadata | Auth identity and contributor listing; generic entity expansion remains tracked by API issue #80 | Deferred to shared entity coverage rather than duplicating an osfr-specific surface |
| Authentication | OSF personal access token for management operations | Token plus username/password fallback with redaction and no credential persistence | Implemented with stronger secret boundary |
| Conflict semantics | High-level R workflows around OSF entities and uploads | Explicit conflict policies, atomic writes, manifests, and conservative destructive confirmation | Implemented with stronger local-write contract |
| Automation surface | R functions and pipe-friendly data frames | JSON CLI and MCP schemas designed for deterministic automation | Implemented for supported OSF scope |

No osfr-specific production gap was accepted. Generic user/entity expansion is
already tracked in [issue #80](https://github.com/edithatogo/osf-cli-go/issues/80),
while the current CLI's explicit transfer and authentication contracts provide
stronger automation safety than adopting R-specific abstractions.
