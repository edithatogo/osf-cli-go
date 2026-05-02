# Plan: MCP Boundary Preparation

## Phase 1: Boundary Audit

- [x] Task: Audit CLI/API/auth/download package coupling and record reusable candidates
- [x] Task: Add or update tests that protect package boundaries before moving code
  - Current reusable seams are already isolated: `internal/auth` owns token lookup/redaction, `internal/osfapi` owns the typed read-only API client, and `internal/download` stays separate from CLI presentation.
  - Added a CLI contract test that keeps the current command surface explicit.

## Phase 2: Minimal Refactor

- [x] Task: Extract small CLI-independent interfaces or helpers only where justified
  - No production refactor was needed; the existing `auth.Source` and `osfapi.Client` seams already serve the intended boundary.
- [x] Task: Keep Cobra command construction and terminal rendering out of reusable boundaries
  - Verified by code review and the added CLI contract guard.

## Phase 3: Roadmap And Review

- [x] Task: Update `docs/mcp-roadmap.md` with concrete next MCP-server milestones
- [x] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
