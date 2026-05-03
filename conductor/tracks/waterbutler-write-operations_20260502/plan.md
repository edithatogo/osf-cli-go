# Plan: WaterButler Write Operations

## Phase 1: WaterButler Upload Endpoint

- [ ] Task: Add `WaterButlerUpload` method to the OSF API client that streams a file via PUT
- [ ] Task: Handle content-type detection and filename encoding for the upload URL
- [ ] Task: Add fixture-backed tests for upload success, 4xx, and 5xx responses
- [ ] Task: Add `osf files upload <file> <node-id>` CLI command with progress output

## Phase 2: Folder Creation

- [ ] Task: Add `WaterButlerCreateFolder` method sending PUT with `?kind=folder`
- [ ] Task: Handle folder path prefix creation (nested folders)
- [ ] Task: Add fixture-backed tests for folder creation and conflict errors
- [ ] Task: Add `osf files mkdir <path> <node-id>` CLI command

## Phase 3: File Deletion

- [ ] Task: Add `WaterButlerDelete` method sending DELETE to the file URL
- [ ] Task: Implement confirmation prompt for the CLI command
- [ ] Task: Add fixture-backed tests for delete success and not-found errors
- [ ] Task: Add `osf files rm <file-id|path> <node-id>` CLI command
- [ ] Task: Run quality gates and `$conductor-review`, apply fixes, and write phase review evidence
