# Spec: Homebrew Tap Adoption

## Overview

Make OSF CLI Go installable through Homebrew using a tap/formula path suitable
for macOS and Linux users.

## Functional Requirements

- Decide whether to use a project-owned tap, GoReleaser-generated tap, or
  upstream Homebrew submission if eligible.
- Generate/validate formula for `osf` and document `osf-mcp` availability.
- Validate checksums, install paths, completions, version output, and uninstall.
- Use Chrome only for GitHub/browser auth if repository setup or PR submission
  requires it; ask the user to log in if blocked.
- Record tap URL, PR/release evidence, validation output, and blockers.

## Acceptance Criteria

- `brew install` path is available or exact blocker is recorded.
- Formula validation, local build, Go tests, vet, anti-stub, and review checks pass.
- Installation docs are updated with Homebrew commands.

## Out Of Scope

- Publishing untagged or unsigned release artifacts.
