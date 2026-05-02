# Spec: Files Download CLI

Expose the existing safe download internals through a user-facing `osf files download` command.

## Outcomes

- A user can download one OSF Storage file by file id, file API URL, or file listing record where a download URL is available.
- A user can download a folder tree from a project or component into a local directory with conservative conflict behavior.
- Default behavior never overwrites an existing local file.
- JSON output returns a stable manifest suitable for automation.
- Human output summarizes written, skipped, and failed records without printing secrets.

## Non-Goals

- Upload, sync, delete, or live write operations.
- Background parallel download scheduling unless the command shape requires it.
- Live OSF tests in normal unit-test runs.
