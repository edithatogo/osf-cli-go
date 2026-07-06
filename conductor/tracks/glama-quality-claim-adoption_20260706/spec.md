# Spec: Glama Quality Claim Adoption

## Overview

Get the OSF MCP server indexed, claimed, or improved on Glama and close quality
score gaps while preserving the repo's security and status-honesty rules.

## Functional Requirements

- Search Glama for existing listings and owner-claim/update paths.
- Capture quality, maintenance, install, auth, and metadata warnings.
- Improve repo metadata, docs, MCPB/server metadata, or validation where Glama
  scoring identifies actionable gaps.
- Use Chrome for OAuth/claim flows; ask the user to log in if automated access
  cannot proceed.
- Record final score and explain any remaining non-actionable gaps.

## Acceptance Criteria

- Glama listing is claimed/submitted or blocker evidence is recorded.
- Score is improved as close to 100% as permitted by target constraints.
- All repo-local validation gates pass.

## Out Of Scope

- Adding unsafe remote execution behavior to improve a score.
- Publishing credentials or browser session artifacts.
