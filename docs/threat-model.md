# Threat Model

## Assets

- OSF tokens, usernames, passwords, and Zenodo scoped tokens
- private project identifiers and research files
- local destination paths and downloaded files
- release artifacts, signing identity, and registry metadata

## Threats and controls

| Threat | Control | Evidence |
|---|---|---|
| Credential disclosure | Environment-based credentials, redacted errors, no project-local persistence | `internal/auth`; redaction tests |
| Unsafe local writes | Destination containment, atomic writes, conflict policy, manifests | `internal/download`; path and download tests |
| Accidental OSF mutation | Explicit CLI write commands, confirmation for destructive actions, MCP read-only boundary | command tests; MCP tool contract |
| Supply-chain tampering | Dependency review, CodeQL, vulnerability scan, SBOM, provenance, keyless signing | GitHub security workflows; release evidence |
| Malicious registry configuration | Version and command validation for packaged integrations | `tools/checkreleasecontract` |
| Service or rate-limit failure | Context cancellation, request timeout, typed API errors, opt-in live tests | API/client tests and live matrix |
| Cross-provider confused deputy | Provider-qualified identities, separate credentials, explicit destination capability checks, and sandbox-only writes | repository contract; cross-provider tests |
| Cross-provider data-residency surprise | Dry-run mapping, explicit source/destination provenance, no implicit publication, and retained native metadata | mapping report and provenance-sidecar tests |
| Orphaned sandbox resources | Cancellation-resistant compensation, machine-checked resource disposition, and explicit retained publication URLs | provider validation manifest and sandbox evidence |
| Sandbox claim promoted to production | Separate validation levels, digest-bound evidence, and required public production receipt | `tools/checkproviderrelease` tests |
| Token reused across providers or publication boundary | Distinct environment secrets, protected environments, scoped one-use credentials, and redacted events | provider validation workflow; live evidence |

Residual risk remains for live OSF service behavior, Zenodo production writes,
third-party storage providers, retained irreversible sandbox publications, and
external marketplace review. Those gates require disposable resources,
separately approved credentials, or provider action and are recorded as
blockers or explicit waivers until run.
