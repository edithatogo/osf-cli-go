# pyosf synchronization parity

Last reviewed: 2026-07-14

This comparison uses the public [psychopy/pyosf repository](https://github.com/psychopy/pyosf), its [README](https://github.com/psychopy/pyosf/blob/master/README.md), and the current OSF CLI Go synchronization contracts.

## Source maturity

| Signal | pyosf | OSF CLI Go |
|---|---|---|
| License | MIT | Apache-2.0 |
| Maintenance | 130 commits, 9 stars, 5 forks, and one release; latest shown as v1.0.5 from 2017-02-13 | Active cross-platform Go CLI/MCP project with release, security, registry, test, race, and vet gates |
| Scope | Pure Python library for simple instructed synchronization of one OSF project | Explicit one-shot tree transfer today; continuous synchronization is separately benchmarked and deferred |
| State model | Local `.proj` file plus flat JSON change state; token cache at `~/.pyosf/tokens.json` | No project-local credential/state persistence; manifests and deterministic transfer results are explicit outputs |
| Workflow | Search/login, open project, calculate local/remote changes, apply changes, save project state | Authenticated explicit file/tree download and upload with conflict policy, path safety, and no background writes |
| Compatibility | Python 2.7 and Python 3.4+ claim, Travis and Codecov badges | Go cross-platform builds, offline fixtures, deterministic CLI/MCP tests, anti-stub, race, vet, and release gates |

## Capability comparison

| Capability | pyosf reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| Project discovery | Search users/projects and open a project | Projects list/get, search, and authenticated identity surfaces | Implemented |
| One-shot transfer | `Project.get_changes()` then `changes.apply()` | Explicit `files download --tree` and `files upload` operations with conflicts and manifests | Implemented with stronger explicit boundary |
| Continuous synchronization | Explicitly out of scope; pyosf is instructed basic sync, while osf-sync provides continuous sync | No background sync; OSF Sync parity already defers journal, locking, resume, and conflict design to issue #13 | Deliberately deferred to the shared sync contract |
| Local state | `.proj` project file stores username and sync root; token cache stores auth token in readable JSON | No project-local credential persistence; manifests are transfer evidence rather than hidden mutable state | Rejected for security and reproducibility |
| Change resolution | Flat JSON state and simple resolution rules | Explicit conflict policies and atomic writes; no implicit reconciliation | Implemented with stronger safety controls |

No pyosf-specific production gap was accepted. The useful one-shot transfer
workflow is already covered, while continuous sync belongs to the shared issue #13
contract and pyosf's readable token cache is incompatible with this repository's
credential boundary.
