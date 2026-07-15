# Implementation Plan

## Phase 1: Candidate build

- [x] Task: Select and tag an immutable release candidate commit. (5a1004c)
- [x] Task: Generate all binary, container, MCPB, plugin, checksum, SBOM, and provenance outputs. (5a1004c)
- [x] Task: Conductor - User Manual Verification 'Candidate build' (Protocol in workflow.md) (5a1004c)

## Phase 2: Independent verification

- [x] Task: Verify signatures, attestations, checksums, SBOMs, OCI digests, and package contents in a clean environment. (5a1004c)
- [x] Task: Validate supported-platform installation and dynamic versioning. (5a1004c)
- [x] Task: Review workflow permissions, pinned actions, artifact retention, and release metadata. (5a1004c)
- [x] Task: Conductor - User Manual Verification 'Independent verification' (Protocol in workflow.md) (5a1004c)

## Phase 3: Launch decision

- [x] Task: Publish dated evidence and update the release checklist and launch review. (5a1004c)
- [x] Task: Block the v1.0 tag until every required verification passes or has an explicit decision. (5a1004c)
- [x] Task: Conductor - User Manual Verification 'Launch decision' (Protocol in workflow.md) (5a1004c)
