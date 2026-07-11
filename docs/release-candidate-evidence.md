# Release Candidate Evidence

This file is the required index for a future `v1.0.0-rc.1`. It is intentionally
not a claim that the 1.0 release gate is currently satisfied.

## Local gates

- [x] Unit tests, race tests, and vet pass on the current default branch.
- [x] Registry, review, and anti-stub checks pass.
- [x] MCPB and agent plugin archives build locally.
- [x] Official MCP Registry, Smithery, and Glama publication evidence exists.
- [ ] Cross-platform release-candidate artifacts are independently verified.
- [ ] Signed provenance, SBOM, and artifact verification are attached to the
  candidate release.
- [ ] Compatibility and support policies are reviewed and linked from release
  notes.
- [ ] High-risk live-validation matrix rows are run against disposable OSF data.
- [ ] Security review has no unresolved release-blocking findings.

## External gates

- [ ] Claude/Cowork directory review completed.
- [ ] Codex/Cowork public marketplace or equivalent review completed.
- [ ] Copilot marketplace review completed.
- [ ] Gemini and Qwen gallery/discovery review completed.
- [ ] Remaining registry submissions completed or explicitly waived.

Each unchecked item requires dated evidence or a written waiver in the release
review before a `v1.0.0` tag is created.
