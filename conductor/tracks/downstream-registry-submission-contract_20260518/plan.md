# Plan: Downstream Registry Submission Contract

## Phase 1: Contract And Packetization

- [x] Task: Create a downstream submission contract covering Smithery,
  MCP.Directory, Glama, PulseMCP, Claude directory/Cowork, Codex, Gemini CLI,
  and Qwen Code.
- [x] Task: Record per-target status as published, prepared, or blocked with a
  required receipt/evidence field.
- [x] Task: Prepare directory submission packets with repository URL, registry
  URL, package URL, short description, long description, install command, and
  support/privacy links.

## Phase 2: MCPB Bundle Route

- [x] Task: Add MCPB manifest metadata for the OSF stdio MCP server.
- [x] Task: Add a local script to build a binary-backed MCPB bundle.
- [x] Task: Add a CI workflow to build and upload MCPB bundle artifacts for
  release/manual distribution.
- [~] Task: Validate the manifest and document validation commands.

## Phase 3: Client Plugin Distribution

- [x] Task: Add install and validation instructions for Claude Code/Cowork.
- [x] Task: Add install and validation instructions for Codex.
- [x] Task: Add install and validation instructions for Gemini CLI.
- [x] Task: Add install and validation instructions for Qwen Code.
- [x] Task: Record what is ready for GitHub Releases versus what requires
  external marketplace review.

## Phase 4: Submission Attempts And Evidence

- [ ] Task: Submit or trigger every non-interactive registry/API/CI path that is
  safe and available.
- [x] Task: Record manual web-form submission blockers and exact next actions.
- [~] Task: Record final receipts, workflow run URLs, and verification outputs.

## Phase 5: Review

- [ ] Task: Run JSON/YAML/script validation plus Go checks affected by the new
  packaging work.
- [ ] Task: Update phase-review evidence and reconcile any stale track text.
