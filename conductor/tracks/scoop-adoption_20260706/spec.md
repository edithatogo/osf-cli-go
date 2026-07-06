# Spec: Scoop Adoption

## Overview

Prepare OSF CLI Go for Scoop installation through a project bucket, Scoop
Extras, or a submission packet, depending on eligibility.

## Functional Requirements

- Verify Scoop bucket requirements, manifest fields, release URLs, hashes,
  binary names, autoupdate metadata, and checkver behavior.
- Generate/validate Scoop manifest for `osf` and document `osf-mcp` packaging.
- Use Chrome only for GitHub/browser auth if PR submission requires it; ask the
  user to log in if blocked.
- Record PR URL, bucket URL, validation output, or blocker.

## Acceptance Criteria

- Scoop install route is published/submitted or blocker is precise.
- Manifest validation, install smoke, Go checks, vet, anti-stub, and review pass.
- Windows install docs are updated when route is ready.

## Out Of Scope

- Main bucket submission if project eligibility clearly points to Extras or a
  project-owned bucket.
