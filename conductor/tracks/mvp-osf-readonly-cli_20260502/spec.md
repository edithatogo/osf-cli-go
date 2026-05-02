# MVP OSF Read-Only CLI

## Objective

Create a useful first release of `osf` that can authenticate with OSF, inspect accessible projects/components, list OSF Storage files, and download content safely.

## User Outcomes

- A user can verify which OSF account/token is active.
- A user can list accessible projects and components.
- A user can inspect one OSF node by GUID.
- A user can list files and folders for a node.
- A user can download public or authorized files without accidental overwrite.
- A user can choose human table output or JSON output.

## Scope

- OSF API v2 client with pagination.
- Personal access token support through environment variables.
- Read-only commands for users, nodes, contributors, and files.
- Download implementation for OSF Storage.
- Offline unit tests using fixture responses.

## Out Of Scope

- Project creation.
- File upload.
- Metadata mutation.
- PDF export.
- Shell completions and release packaging.

## Design Notes

- Follow OSF terminology: projects and components are OSF nodes.
- Treat GUIDs and API URLs as valid user inputs where practical.
- Do not require authentication for public project inspection.
- Use conservative filesystem conflict behavior: fail by default when the local destination exists.
