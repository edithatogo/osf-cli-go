# Live OSF Validation Evidence

- Generated: 2026-05-02T00:00:00Z
- Mode: dry-run
- Historical note: this file records the original 2026-05-02 dry-run before
  `files download` landed. Current closeout evidence is recorded in
  `phase-review.md`; the command now exists and is included in the opt-in live
  validation tool.
- Environment:
  - OSF_TOKEN: missing
  - OSF_VALIDATE_PROJECT: missing
  - OSF_LIVE_VALIDATION: false
- Planned coverage:
  - auth whoami: planned
  - projects list: planned
  - projects get: planned
  - components list: planned
  - files list: planned
  - files download: historical pre-download state; now implemented
- Results:
  - auth whoami: planned
    - Output: not executed in dry-run mode
  - projects list: planned
    - Output: not executed in dry-run mode
  - projects get: planned
    - Output: not executed in dry-run mode
  - components list: planned
    - Output: not executed in dry-run mode
  - files list: planned
    - Output: not executed in dry-run mode
  - files download: historical pre-download state
    - Output: superseded by current `tools/livevalidation` coverage after the
      download command landed
