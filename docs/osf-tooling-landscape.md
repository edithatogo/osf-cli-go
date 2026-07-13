# OSF tooling landscape

Last reviewed: 2026-07-13

This inventory tracks public repositories that expose substantial reusable OSF
functionality. It excludes one-off download scripts, demos, deprecated mirrors,
and repositories that merely mention the OSF API.

| Repository | Distinct area to assess | Roadmap |
|---|---|---|
| [SourceShift/osf-mcp-server](https://github.com/SourceShift/osf-mcp-server) | Direct MCP competitor; search, registrations, preprints, and file management | [#8](https://github.com/edithatogo/osf-cli-go/issues/8) closed; dated comparison evidence in the archived SourceShift track |
| [osfclient/osfclient](https://github.com/osfclient/osfclient) | Mature Python library and CLI for OSF Storage | [#9](https://github.com/edithatogo/osf-cli-go/issues/9) |
| [ropensci/osfr](https://github.com/ropensci/osfr) | Broad R entity coverage and explicit conflict behavior | [#10](https://github.com/edithatogo/osf-cli-go/issues/10) |
| [psychopy/pyosf](https://github.com/psychopy/pyosf) | Project synchronization workflows | [#11](https://github.com/edithatogo/osf-cli-go/issues/11) |
| [datalad/datalad-osf](https://github.com/datalad/datalad-osf) | DataLad datasets, remotes, and annex-oriented workflows | [#12](https://github.com/edithatogo/osf-cli-go/issues/12) |
| [CenterForOpenScience/osf-sync](https://github.com/CenterForOpenScience/osf-sync) | Desktop synchronization semantics | [#13](https://github.com/edithatogo/osf-cli-go/issues/13) |
| [J535D165/datahugger](https://github.com/J535D165/datahugger) | DOI-oriented cross-repository retrieval | [#14](https://github.com/edithatogo/osf-cli-go/issues/14) |
| [mims-harvard/ToolUniverse](https://github.com/mims-harvard/ToolUniverse) | AI-scientist OSF Preprints tools | [#15](https://github.com/edithatogo/osf-cli-go/issues/15) |
| [wentorai/research-plugins](https://github.com/wentorai/research-plugins) | Agent skills for OSF discovery and full text | [#16](https://github.com/edithatogo/osf-cli-go/issues/16) |
| [CoLRev-Environment/colrev](https://github.com/CoLRev-Environment/colrev) | Literature-review ingestion and metadata normalization | [#17](https://github.com/edithatogo/osf-cli-go/issues/17) |
| [jasp-stats/jasp-desktop](https://github.com/jasp-stats/jasp-desktop) | Mature desktop OSF browsing and file workflows | [#18](https://github.com/edithatogo/osf-cli-go/issues/18) |
| [Lucy-Family-Institute/presqt](https://github.com/Lucy-Family-Institute/presqt) | Preservation, transfer, and metadata enrichment | [#19](https://github.com/edithatogo/osf-cli-go/issues/19) |
| [scienceverse/metacheck](https://github.com/scienceverse/metacheck) | Research-output and preregistration checks | [#20](https://github.com/edithatogo/osf-cli-go/issues/20) |

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
