# Implementation Plan

## Phase 1: API and capability mapping

- [ ] Task: Pin official Zenodo API source, terms, retrieval date, sandbox, and endpoint evidence.
- [ ] Task: Map OSF and Zenodo records, metadata, files, identifiers, auth, limits, and publication states.
- [ ] Task: Define shared provider capability vocabulary and explicit non-equivalences.
- [ ] Task: Conductor - User Manual Verification 'API and capability mapping' (Protocol in workflow.md)

## Phase 2: Shared provider contract

- [ ] Task: Define typed provider interfaces and domain models without weakening existing OSF contracts.
- [ ] Task: Add fixture-backed Zenodo read-only client tests and redacted error behavior.
- [ ] Task: Add provider-neutral transfer, pagination, retry, cancellation, checksum, and rate-limit contracts.
- [ ] Task: Conductor - User Manual Verification 'Shared provider contract' (Protocol in workflow.md)

## Phase 3: CLI and MCP read-only parity

- [ ] Task: Add read-only Zenodo discovery behind an explicit provider boundary.
- [ ] Task: Add CLI/MCP compatibility fixtures and documentation for supported and deferred operations.
- [ ] Task: Validate OSF behavior remains unchanged and no unsupported Zenodo writes are advertised.
- [ ] Task: Conductor - User Manual Verification 'CLI and MCP read-only parity' (Protocol in workflow.md)

## Phase 4: Transfers and publication workflows

- [ ] Task: Design sandbox-first upload, download, DOI reservation, publish, discard, and cleanup workflows.
- [ ] Task: Validate provider-specific checksums, limits, conflicts, resumability, and cancellation.
- [ ] Task: Add explicit authorization and confirmation gates for publication and destructive actions.
- [ ] Task: Conductor - User Manual Verification 'Transfers and publication workflows' (Protocol in workflow.md)

## Phase 5: Release integration

- [ ] Task: Add provider contract checks to CI, release validation, feature matrix, and launch review.
- [ ] Task: Record live/sandbox evidence or a dated external-validation blocker.
- [ ] Task: Conductor - User Manual Verification 'Release integration' (Protocol in workflow.md)
