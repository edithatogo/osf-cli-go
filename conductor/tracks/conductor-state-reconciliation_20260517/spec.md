# Spec: Conductor State Reconciliation

## Problem

The project track index and per-track plans show all current tracks complete,
but a repo-level status scan found older phase-review prose that still described
some completed tracks as pending or open. That makes the Conductor status harder
to trust during handoff.

## Goals

- Verify there are no unchecked tasks in per-track `plan.md` files.
- Preserve historical evidence while correcting stale status language that now
  conflicts with the current track index and plan state.
- Record the reconciliation as its own granular track so the cleanup is auditable.
- Avoid changing implementation behavior or running live OSF write operations.

## Non-Goals

- Do not rewrite historical design tracks that intentionally use scaffolded or
  roadmap language.
- Do not mark write-operation live validation as complete without an explicit
  disposable OSF project and user approval.
- Do not alter user credentials or credential storage.

## Acceptance Criteria

- `conductor/tracks.md` includes this reconciliation track.
- This track has a checked plan and phase review.
- Stale completion-status language in completed delivery tracks is updated with
  current evidence.
- `git diff --check` passes after the documentation edits.
