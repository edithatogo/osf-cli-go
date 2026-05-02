# Download Safety

## Objective

Implement safe file and folder downloads from OSF Storage with conservative conflict behavior.

## Acceptance Criteria

- Downloads stream to a temporary file and rename only after success.
- Existing local files fail by default.
- Explicit conflict policies support fail, skip, and overwrite.
- Remote paths cannot escape the selected destination directory.
- Folder downloads can emit a manifest.
