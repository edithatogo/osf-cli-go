# Spec: Downstream Registry Submission Contract

## Objective

Finish the non-official-registry distribution requirements left after the
official MCP Registry publication: Smithery/MCPB packaging, downstream directory
submission packets, and client-plugin distribution contracts.

## Contract

The repo must distinguish between:

- **Published**: a live registry or package endpoint was submitted and verified.
- **Prepared**: repo-local metadata, package, workflow, or submission packet is
  ready, but a live manual form, hosted endpoint, or external review remains.
- **Blocked**: the target requires a credential, UI-only form, hosted endpoint,
  or package type that does not yet exist.

No downstream target may be marked published without a receipt URL, API response,
workflow run, or equivalent evidence in the track review.

## Target Surfaces

- Smithery via MCPB bundle first; hosted Streamable HTTP remains future work.
- MCP.Directory, Glama, PulseMCP, and similar discovery directories.
- Claude plugin directory and Cowork/org plugin deployment.
- Codex, Gemini CLI, and Qwen Code extension/package distribution.
- GitHub Releases as the transport for generated binary/MCPB artifacts.

## Acceptance Criteria

- A repo-local submission contract lists every target, live status, submission
  method, required artifact, and blocker.
- MCPB bundle metadata exists and can be built by a documented command and CI.
- Directory submission packets are prepared with exact copy/paste fields and
  links to evidence.
- Client plugin distribution docs explain install, validation, and bundled
  binary expectations.
- External submissions are attempted only where a non-interactive CLI/API or
  approved authenticated workflow exists; manual web-form blockers are recorded
  precisely.

## Non-Goals

- Do not build a public Streamable HTTP service in this track.
- Do not add write-capable MCP tools.
- Do not store OSF tokens, GitHub tokens, registry tokens, signing keys, or
  submission credentials in the repository.
