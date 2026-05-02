# Spec: MCP Boundary Preparation

Prepare minimal reusable boundaries for a future MCP server without compromising CLI-first design.

## Outcomes

- Identify CLI-independent service interfaces for OSF API access, auth, and downloads.
- Move only genuinely reusable code when it improves testability or avoids future duplication.
- Keep command presentation, Cobra wiring, and terminal output inside CLI packages.
- Update the MCP roadmap with concrete future package/API decisions.

## Non-Goals

- Implementing an MCP server.
- Adding MCP dependencies.
- Promoting unstable public packages before the CLI contract is proven.
