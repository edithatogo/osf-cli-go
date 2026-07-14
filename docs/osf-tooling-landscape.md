# OSF tooling landscape

Last reviewed: 2026-07-14

This inventory tracks public repositories that expose substantial reusable OSF
functionality. It excludes one-off download scripts, demos, deprecated mirrors,
and repositories that merely mention the OSF API.

| Repository | Distinct area to assess | Roadmap |
|---|---|---|
| [SourceShift/osf-mcp-server](https://github.com/SourceShift/osf-mcp-server) | Direct MCP competitor; search, registrations, preprints, and file management | [#8](https://github.com/edithatogo/osf-cli-go/issues/8) closed; dated comparison evidence in the archived SourceShift track |
| [osfclient/osfclient](https://github.com/osfclient/osfclient) | BSD-3-Clause Python library and CLI for OSF Storage with listing, fetch, clone, upload, and URL workflows | [#9](https://github.com/edithatogo/osf-cli-go/issues/9); dated comparison evidence in [`docs/osfclient-cli-parity.md`](osfclient-cli-parity.md) |
| [ropensci/osfr](https://github.com/ropensci/osfr) | MIT R interface with node, file/folder, user, download, upload, and component workflows | [#10](https://github.com/edithatogo/osf-cli-go/issues/10); dated comparison evidence in [`docs/osfr-entity-parity.md`](osfr-entity-parity.md) |
| [psychopy/pyosf](https://github.com/psychopy/pyosf) | MIT Python one-project synchronization with local change state and explicit apply | [#11](https://github.com/edithatogo/osf-cli-go/issues/11); dated comparison evidence in [`docs/pyosf-sync-parity.md`](pyosf-sync-parity.md) |
| [datalad/datalad-osf](https://github.com/datalad/datalad-osf) | DataLad datasets, Git remotes, and git-annex-oriented workflows | [#12](https://github.com/edithatogo/osf-cli-go/issues/12) closed; dated comparison evidence in the archived DataLad track |
| [CenterForOpenScience/osf-sync](https://github.com/CenterForOpenScience/osf-sync) | LGPL-3.0 OSX/Windows desktop synchronization; no published GitHub releases in the reviewed snapshot | [#13](https://github.com/edithatogo/osf-cli-go/issues/13); parity evidence in [`docs/osf-sync-parity.md`](osf-sync-parity.md) |
| [J535D165/datahugger](https://github.com/J535D165/datahugger) | DOI-oriented cross-repository retrieval | [#14](https://github.com/edithatogo/osf-cli-go/issues/14) closed; dated comparison evidence in the archived Datahugger track |
| [mims-harvard/ToolUniverse](https://github.com/mims-harvard/ToolUniverse) | AI-scientist OSF Preprints search and literature-agent ecosystem | [#15](https://github.com/edithatogo/osf-cli-go/issues/15); dated comparison evidence in the archived ToolUniverse track |
| [wentorai/research-plugins](https://github.com/wentorai/research-plugins) | MIT research-agent plugin with 433 skills, 34 scholarly tools, and an OSF Preprints module | [#16](https://github.com/edithatogo/osf-cli-go/issues/16); parity evidence in [`docs/research-plugins-osf-parity.md`](research-plugins-osf-parity.md) |
| [matsjfunke/paperclip](https://github.com/matsjfunke/paperclip) | Archived Python FastMCP server for multi-provider metadata and PDF-to-Markdown retrieval, including OSF | [#45](https://github.com/edithatogo/osf-cli-go/issues/45); dated comparison evidence in [`docs/paperclip-research-retrieval-parity.md`](paperclip-research-retrieval-parity.md) |
| [CoLRev-Environment/colrev](https://github.com/CoLRev-Environment/colrev) | Literature-review ingestion, OSF search, and metadata normalization | [#17](https://github.com/edithatogo/osf-cli-go/issues/17); dated comparison evidence in the archived CoLRev track |
| [jasp-stats/jasp-desktop](https://github.com/jasp-stats/jasp-desktop) | Mature desktop OSF browsing, authenticated file workflows, folder management, and provider checksum awareness | [#18](https://github.com/edithatogo/osf-cli-go/issues/18); dated comparison evidence in the JASP track |
| [Lucy-Family-Institute/presqt](https://github.com/Lucy-Family-Institute/presqt) | Preservation, transfer, and metadata enrichment | [#19](https://github.com/edithatogo/osf-cli-go/issues/19) |
| [scienceverse/metacheck](https://github.com/scienceverse/metacheck) | Experimental R research-output checks, OSF retrieval, and preregistration-oriented modules | [#20](https://github.com/edithatogo/osf-cli-go/issues/20); dated comparison evidence in the Metacheck track |

## Recommended improvements

The highest-value improvement themes are:

1. Expand MCP coverage to the CLI's search, preprint, registration, export, and
   safe file-write capabilities with explicit read/write capability metadata.
2. Add resumable, manifest-driven bidirectional synchronization with dry-run,
   conflict detection, checksums, and machine-readable plans.
3. Add DOI and URL resolution so OSF content can participate in broader
   research-data retrieval workflows.
4. Add metadata, FAIRness, and preregistration validation with stable findings
   schemas suitable for CI and MCP clients.
5. Provide agent-oriented discovery and full-text workflows while retaining
   provenance, pagination, and deterministic output.
6. Add interoperability contracts for DataLad, JASP, review tools, and
   preservation/transfer systems rather than embedding their internal models.

Each recommendation is gated by the source-specific comparison tracks. A
feature is adopted only when it improves the product without weakening safety,
portability, testability, or licensing clarity.
