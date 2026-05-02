# Plan: Release Packaging

## Phase 1: Versioned Build

- [x] Task: Add version metadata variables and ldflags-compatible build wiring
- [x] Task: Add a local build script or Make target that creates a binary artifact

## Phase 2: CLI Distribution Surface

- [x] Task: Add Cobra completion commands for supported shells
- [x] Task: Add release-check documentation and update examples

## Phase 3: Release Automation

- [x] Task: Add conservative GoReleaser or equivalent config with publishing disabled by default
- [x] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
