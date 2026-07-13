# DataLad OSF comparison

Reviewed: 2026-07-13

## Evidence

- Repository: [datalad/datalad-osf](https://github.com/datalad/datalad-osf)
- README and installation workflow: [README.md](https://github.com/datalad/datalad-osf/blob/main/README.md)
- Package entry points: [setup.py](https://github.com/datalad/datalad-osf/blob/main/setup.py)
- Source workflows: [`create-sibling-osf`](https://github.com/datalad/datalad-osf/blob/main/datalad_osf/create_sibling_osf.py), [`git-remote-osf`](https://github.com/datalad/datalad-osf/blob/main/datalad_osf/git_remote.py), and [`git-annex-remote-osf`](https://github.com/datalad/datalad-osf/blob/main/datalad_osf/annex_remote.py)
- GitHub metadata checked on 2026-07-13: default branch `main`, latest push 2024-10-02, 15 stars, 14 open issues, and no machine-readable SPDX license assertion (`NOASSERTION`). The repository's source `LICENSE` is retained as the authoritative license artifact.
- Latest tagged release is `0.3.0`, published 2023-06-09. The repository has a documentation workflow but no current release workflow comparable to this repository's release-contract gates.

## Capability comparison

| Capability | DataLad OSF | OSF CLI Go | Decision |
|---|---|---|---|
| Public project/file retrieval | DataLad clone and git-annex remotes over `osf://` | `projects`, `files download`, and `export` with bounded, safe local writes | Covered for general CLI workflows; DataLad protocol remains separate |
| OSF project creation | `create-sibling-osf` creates a node and configures a DataLad sibling | CLI supports explicit project creation with confirmation | Covered at the OSF API layer; do not silently configure Git/DataLad state |
| Annex-backed storage | `git-annex-remote-osf` supports annex and export modes | No git-annex dependency or special-remote protocol | Defer to [#69](https://github.com/edithatogo/osf-cli-go/issues/69); requires an interoperability track with git-annex fixtures and protocol ownership |
| Git remote helper | `git-remote-osf` supports `osf://` repository URLs | No Git remote helper | Defer to [#69](https://github.com/edithatogo/osf-cli-go/issues/69); outside the CLI/API/MCP boundary and requires Git transport integration |
| Credential manager integration | DataLad credential manager with token or username/password | Environment-backed token and username/password fallback with redaction | Reject persistence in this tool; preserve stateless automation and secret-safety contract |
| Dataset metadata/tags/categories | DataLad sibling setup supplies dataset-oriented metadata | Project create/update exposes OSF node metadata | Covered where OSF API semantics are shared; DataLad-specific defaults deferred |
| Transfer conflict/update behavior | Annex and export modes handle repeated pushes and force operations | Explicit conflict policies, manifests, path protection, and conservative writes | OSF CLI Go exceeds general local-write safety; annex semantics deferred |
| Tests and release maturity | Pytest/annex integration and Windows coverage are documented; latest release 2023 | Deterministic Go tests, race/vet/lint/vulnerability/registry/release gates | OSF CLI Go exceeds repository-level release evidence |

## Scope result

DataLad OSF's distinctive value is its Git/DataLad/git-annex integration, not a
missing ordinary OSF API endpoint. The current CLI already covers the safe
general-purpose project, file, export, authentication, and write primitives.
The annex remote and Git remote helper are deliberately deferred to [issue
#69](https://github.com/edithatogo/osf-cli-go/issues/69) because implementing
them would require a separate protocol contract, git-annex and DataLad
integration fixtures, and platform-specific validation. No credentials or
live OSF writes are needed for this benchmark.
