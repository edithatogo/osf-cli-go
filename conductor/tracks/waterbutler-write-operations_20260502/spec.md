# WaterButler Write Operations

## Objective

Add file upload, folder creation, and file deletion via the OSF WaterButler storage API. These are the biggest missing pieces for a complete CLI.

## Acceptance Criteria

- File upload streams content via PUT to the WaterButler files provider URL
- Folder creation sends PUT with `?kind=folder` query parameter
- File deletion sends DELETE request to the WaterButler file URL
- All operations handle 4xx/5xx responses with actionable error messages
- All operations are covered by fixture-backed unit tests

## Non-Goals

- Recursive directory upload
- Batch delete operations
- File move or rename
- Sync or rsync-style operations
