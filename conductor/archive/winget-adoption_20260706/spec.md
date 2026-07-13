# Spec: WinGet Adoption

## Overview

Prepare and submit OSF CLI Go to Windows Package Manager (`winget`) using
release artifacts and manifests that meet current repository validation.

## Functional Requirements

- Verify WinGet package requirements, package identifier, installer type,
  release artifact URLs, checksums, license, tags, and publisher fields.
- Generate/update WinGet manifest files or a submission PR packet.
- Validate manifests using available WinGet tooling or schema checks.
- Use Chrome only for GitHub/browser auth if CLI submission is blocked; ask
  user to log in when needed.
- Record PR URL, validation logs, review status, or blocker.

## Acceptance Criteria

- WinGet PR/submission is created or blocker is precise.
- Manifest validation and repo validation gates pass.
- Windows install docs mention WinGet when accepted or prepared.

## Out Of Scope

- Publishing unsigned or unreleased binaries as stable installers.
