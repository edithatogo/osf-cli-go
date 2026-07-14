# osfclient CLI parity

Last reviewed: 2026-07-14

This comparison uses the public [osfclient/osfclient repository](https://github.com/osfclient/osfclient), its [README](https://github.com/osfclient/osfclient/blob/master/README.rst), and the current OSF CLI Go contracts.

## Source maturity

| Signal | osfclient | OSF CLI Go |
|---|---|---|
| License | BSD-3-Clause | Apache-2.0 |
| Maintenance | 306 commits, 145 stars, 57 forks, and 6 tags in the reviewed snapshot | Active versioned CLI/MCP project with release, security, registry, test, race, and vet gates |
| Scope | Python library and CLI focused on OSF file storage | Go CLI and stdio MCP server covering OSF projects, search, preprints, metadata, storage, auth, export, and validation |
| Authentication | `OSF_TOKEN` or username/password; README documents `.osfcli.config` credential storage | Token or username/password fallback with redaction and no project-local credential persistence |
| Transfer workflows | `ls`, `fetch`, `clone`, `upload`, and `geturl` | `files list`, explicit `files download --file/--tree`, `files upload`, and structured file links |
| Testing and packaging | Python test/development configuration and PyPI installation | Offline fixtures, deterministic CLI/MCP tests, cross-platform release artifacts, and security/release gates |

## Capability comparison

| Capability | osfclient reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| Project/file listing | `osf ls` | `files list` with structured output and provider traversal | Implemented |
| Individual retrieval | `osf fetch remote/path local/path` | Explicit `files download --file` with atomic writes, conflict policy, manifests, and path safety | Implemented with stronger safety controls |
| Recursive retrieval | `osf clone` | Explicit `files download --tree`; no implicit bulk write | Implemented with explicit safety boundary |
| Upload | `osf upload`, including recursive upload | `files upload` with explicit authentication and conflict behavior | Implemented |
| File URL access | `osf geturl` | File metadata exposes OSF links; `open` supports browser navigation | Implemented |
| Authentication | Environment variables or `.osfcli.config` | Redacted token/password fallback without project-local secret persistence | Deliberately safer; no config-file credential compatibility |
| Initialization | `osf init` writes local project configuration | Explicit command arguments and environment-based configuration | Rejected to preserve portability and secret boundaries |

No production gap was accepted. The source's storage-focused workflows are covered
by existing commands, while its implicit clone and local credential-file patterns are
not adopted because they would weaken explicit-write and secret-handling guarantees.
