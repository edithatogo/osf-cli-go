# OSF Sync parity

Last reviewed: 2026-07-14

This comparison uses the public [Center for Open Science `osf-sync`
repository](https://github.com/CenterForOpenScience/osf-sync), its
[README](https://github.com/CenterForOpenScience/osf-sync/blob/develop/README.md),
and the [OSF API documentation](https://developer.osf.io/). It assesses
user-facing behavior and maintenance signals, not copied implementation.

## Source maturity

| Signal | OSF Sync | OSF CLI Go |
|---|---|---|
| Primary language | Python desktop application | Go CLI and stdio MCP server |
| License | LGPL-3.0 | Apache-2.0 |
| Platforms | OSX and Windows build workflows are advertised | Windows, macOS, and Linux CLI/release workflows |
| Release surface | No GitHub releases published in the reviewed repository snapshot | Versioned binaries, OCI image, MCPB, checksums, SBOM/provenance gates |
| Automation surface | Desktop synchronization application | Non-interactive CLI, JSON output, shell completion, and MCP |
| Maintenance evidence | Public repository with 1,139 commits and 10 stars at review time | Active CI, offline tests, race/vet/security/release gates |

## Capability comparison

| Capability | OSF Sync reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| Authenticate to OSF | Desktop application authentication | `OSF_TOKEN` preferred, username/password fallback, redacted errors | Implemented with a non-persistent automation contract |
| Continuous bidirectional sync | Desktop synchronization for OSF projects | No background daemon or file watcher | Deferred to issue #13; requires a separate journal, locking, resume, conflict, and destructive-write contract |
| One-shot remote-to-local transfer | Implied by desktop sync workflow | `osf files download --file` and `--tree` | Implemented; atomic writes, path checks, conflict policies, and manifests |
| Local-to-remote transfer | Implied by desktop sync workflow | `osf files upload`, `files mkdir`, and `files rm` | Implemented as explicit, authenticated, confirmation-aware commands |
| Conflict handling | Desktop behavior is not specified in the public README | Download `fail|skip|overwrite`; upload defaults to fail and supports explicit overwrite | Implemented conservatively; sync-level three-way merge is deferred |
| Project/file discovery | OSF project synchronization | `projects`, `components`, `files`, `search`, `export`, and MCP read tools | Implemented with stable table/JSON contracts |
| Cross-platform packaging | OSX and Windows desktop builds | Windows/macOS/Linux release artifacts and MCP packages | Implemented for CLI/MCP distribution |
| Offline validation | Public README exposes build badges but no reproducible parity contract | Fixtures, unit tests, race tests, vet, stub, registry, release, and MCP-quality checks | Implemented |

## Deferred sync contract

The continuous-sync advantage is intentionally not implemented in this track.
A production-quality daemon would need an opt-in design for local state
directories, file identity, remote and local tombstones, resumable transfers,
three-way conflict resolution, offline queues, rate limits, process locking,
crash recovery, and explicit handling of remote deletion. Adding a watcher that
merely mirrors files would be unsafe and would not satisfy the parity goal.

The existing one-shot transfer surface is the supported alternative. The
feature is tracked in [issue #13](https://github.com/edithatogo/osf-cli-go/issues/13)
and the maintained [feature matrix](feature-matrix.md); no credentials or live
sync claims are made.
