# Threat Model

## Assets

- OSF tokens, usernames, and passwords
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

Residual risk remains for live OSF service behavior, third-party storage
providers, and external marketplace review. Those gates require disposable
resources or provider action and are recorded as waivers until run.
