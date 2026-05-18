# Plan: WaterButler Write Operations

## Reconciliation Note

Status reconciled on 2026-05-16 against the current codebase. Core API and CLI write operations exist and have offline tests. Follow-up implementation on 2026-05-16 added upload content-type detection, progress wiring, nested path handling, and delete confirmation. Full validation and phase-review evidence were recorded on 2026-05-16.

## Phase 1: WaterButler Upload Endpoint

- [x] Task: Add `UploadFile` method to the OSF API client that streams a file via PUT
- [x] Task: Handle filename encoding for the upload URL
- [x] Task: Detect and set content type for uploads when available
- [x] Task: Add fixture-backed tests for upload success and error responses
- [x] Task: Add `osf files upload <file> <node-id>` CLI command with progress output

## Phase 2: Folder Creation

- [x] Task: Add `CreateFolder` method sending PUT with `?kind=folder`
- [x] Task: Handle folder path prefix creation (nested folders)
- [x] Task: Add fixture-backed tests for folder creation and error responses
- [x] Task: Add `osf files mkdir <path> <node-id>` CLI command

## Phase 3: File Deletion

- [x] Task: Add `DeleteFile` method sending DELETE to the file URL
- [x] Task: Implement confirmation prompt for the CLI command
- [x] Task: Add fixture-backed tests for delete success and not-found errors
- [x] Task: Add `osf files rm <file-id|path> <node-id>` CLI command
- [x] Task: Run quality gates and `$conductor-review`, apply fixes, and write phase review evidence
