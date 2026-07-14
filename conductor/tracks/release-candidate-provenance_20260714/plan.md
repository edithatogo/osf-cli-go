# Implementation Plan

## Phase 1: Candidate build

- [ ] Task: Select and tag an immutable release candidate commit.
- [ ] Task: Generate all binary, container, MCPB, plugin, checksum, SBOM, and provenance outputs.
- [ ] Task: Conductor - User Manual Verification 'Candidate build' (Protocol in workflow.md)

## Phase 2: Independent verification

- [ ] Task: Verify signatures, attestations, checksums, SBOMs, OCI digests, and package contents in a clean environment.
- [ ] Task: Validate supported-platform installation and dynamic versioning.
- [ ] Task: Review workflow permissions, pinned actions, artifact retention, and release metadata.
- [ ] Task: Conductor - User Manual Verification 'Independent verification' (Protocol in workflow.md)

## Phase 3: Launch decision

- [ ] Task: Publish dated evidence and update the release checklist and launch review.
- [ ] Task: Block the v1.0 tag until every required verification passes or has an explicit decision.
- [ ] Task: Conductor - User Manual Verification 'Launch decision' (Protocol in workflow.md)
